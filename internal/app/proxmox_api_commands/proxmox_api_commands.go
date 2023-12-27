package proxmox_api_commands

import (
	"fmt"
	"time"

	"github.com/EgorRepnikov/proxmox-api-commands/internal/pkg/constants"
	"github.com/EgorRepnikov/proxmox-api-commands/internal/pkg/env"
	"github.com/EgorRepnikov/proxmox-api-commands/internal/pkg/http_client"
	"github.com/EgorRepnikov/proxmox-api-commands/internal/pkg/utils"
	"github.com/fasthttp/router"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

type ProxmoxLoginResponse struct {
	Data struct {
		Ticket              string `json:"ticket"`
		CsrfPreventionToken string `json:"CSRFPreventionToken"`
	} `json:"data"`
}

func InitProxmoxApiCommands(router *router.Router, env *env.Env, httpClient *http_client.HttpClient) {
	proxmoxApiCommands := &ProxmoxApiCommands{
		router:     router,
		env:        env,
		httpClient: httpClient,
	}
	router.GET("/proxmox-api-commands/node/{node}/{command}", proxmoxApiCommands.handleNodeCommand)
	router.GET("/proxmox-api-commands/{node}/vm/{vm}/{command}", proxmoxApiCommands.handleVmCommand)
}

type ProxmoxApiCommands struct {
	router     *router.Router
	env        *env.Env
	httpClient *http_client.HttpClient

	loginDateExpiration time.Time
	csrfPreventionToken string
	ticket              string
}

func (p *ProxmoxApiCommands) handleNodeCommand(ctx *fasthttp.RequestCtx) {
	node := ctx.UserValue("node")
	if node == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte("Wrong 'node' param"))
		return
	}

	command := ctx.UserValue("command")
	if _, ok := constants.NODE_COMMANDS[command.(string)]; !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte("Wrong 'command' param"))
		return
	}

	authHeaders, err := p.generateAuthHeaders()
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	status, body, err := p.httpClient.POST(
		fmt.Sprintf("%s/api2/json/nodes/%s/status?command=%s", p.env.PROXMOX_HOST, node, command),
		[]byte{},
		authHeaders,
		p.env.REQUEST_TIMEOUT,
	)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte("Error on request"))
		return
	}

	ctx.SetBody([]byte(fmt.Sprintf("%d\n\n%s", status, string(body))))
}

func (p *ProxmoxApiCommands) handleVmCommand(ctx *fasthttp.RequestCtx) {
	node := ctx.UserValue("node")
	if node == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte("Wrong 'node' param"))
		return
	}

	vm := ctx.UserValue("vm")
	if vm == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte("Wrong 'vm' param"))
		return
	}

	command := ctx.UserValue("command")
	if _, ok := constants.VM_COMMANDS[command.(string)]; !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte("Wrong 'command' param"))
		return
	}

	authHeaders, err := p.generateAuthHeaders()
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	status, body, err := p.httpClient.POST(
		fmt.Sprintf("%s/api2/json/nodes/%s/qemu/%s/status/%s", p.env.PROXMOX_HOST, node, vm, command),
		[]byte{},
		authHeaders,
		p.env.REQUEST_TIMEOUT,
	)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte("Error on request"))
		return
	}

	ctx.SetBody([]byte(fmt.Sprintf("%d\n\n%s", status, string(body))))
}

func (p *ProxmoxApiCommands) generateAuthHeaders() ([]*http_client.RequestHeader, error) {
	err := p.loginIfNeed()
	if err != nil {
		return nil, err
	}

	return []*http_client.RequestHeader{
		{Key: "Cookie", Value: fmt.Sprintf("PVEAuthCookie=%s", p.ticket)},
		{Key: "CSRFPreventionToken", Value: p.csrfPreventionToken},
	}, nil
}

func (p *ProxmoxApiCommands) loginIfNeed() error {
	if p.loginDateExpiration.Before(time.Now()) {
		status, body, err := p.httpClient.POST(
			fmt.Sprintf("%s/api2/json/access/ticket?username=%s&password=%s", p.env.PROXMOX_HOST, p.env.PROXMOX_USERNAME, p.env.PROXMOX_PASSWORD),
			[]byte{},
			[]*http_client.RequestHeader{},
			p.env.REQUEST_TIMEOUT,
		)
		if err != nil {
			log.Debug().Err(err).Str("body", string(body)).Msg("Error on login request")
			return errors.Wrap(err, "Error on login request")
		}
		if status != 200 {
			log.Debug().Str("body", string(body)).Msgf("Error on login request, status=%d", status)
			return errors.Errorf("Error on login request, status=%d", status)
		}
		var proxmoxLoginResponse *ProxmoxLoginResponse
		err = utils.JsonUnmarshal(body, &proxmoxLoginResponse)
		if err != nil {
			log.Debug().Err(err).Str("body", string(body)).Msg("Error on parse login response")
			return errors.Wrap(err, "Error on parse login response")
		}
		p.loginDateExpiration = time.Now().Add(p.env.AUTHORIZATION_TTL)
		p.ticket = proxmoxLoginResponse.Data.Ticket
		p.csrfPreventionToken = proxmoxLoginResponse.Data.CsrfPreventionToken
	}
	return nil
}
