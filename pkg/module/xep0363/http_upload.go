package xep0363

import (
	"context"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/jackal-xmpp/stravaganza"
	"github.com/ortuman/jackal/pkg/cluster/resourcemanager"
	"github.com/ortuman/jackal/pkg/hook"
	"github.com/ortuman/jackal/pkg/router"
	"github.com/ortuman/jackal/pkg/storage/repository"
)

const (
	// ModuleName represents http_upload module name.
	ModuleName = "http_upload"

	// XEPNumber represents mam XEP number.
	XEPNumber = "0363"

	HttpUploadNamespace = "urn:xmpp:http:upload:0"
)

// HttpUpload represents http_upload (XEP-0363) module type.
type HttpUpload struct {
	rep repository.Repository
	//hosts  hosts
	router router.Router
	resMng resourcemanager.Manager
	hk     *hook.Hooks
	logger kitlog.Logger
}

// New returns a new initialized BlockList instance.
func New(
	router router.Router,
	//hosts *host.Hosts,
	resMng resourcemanager.Manager,
	rep repository.Repository,
	hk *hook.Hooks,
	logger kitlog.Logger,
) *HttpUpload {
	return &HttpUpload{
		rep:    rep,
		router: router,
		//hosts:  hosts,
		resMng: resMng,
		hk:     hk,
		logger: kitlog.With(logger, "module", ModuleName, "xep", XEPNumber),
	}
}

func (m *HttpUpload) Start(ctx context.Context) error {
	level.Info(m.logger).Log("msg", "started http_upload module")
	return nil
	// m.hk.AddHook(hook.C2SStreamElementReceived, m.onElementRecv, hook.DefaultPriority)
}

func (m *HttpUpload) Stop(ctx context.Context) error {
	level.Info(m.logger).Log("msg", "stopped http_upload module")
	return nil
	// m.hk.RemoveHook(hook.C2SStreamElementReceived, m.onElementRecv)
}

func (m *HttpUpload) ServerFeatures(_ context.Context) ([]string, error) {
	return []string{HttpUploadNamespace}, nil
}

// AccountFeatures returns account last activity features.
func (m *HttpUpload) AccountFeatures(_ context.Context) ([]string, error) {
	return nil, nil
}

// Name returns roster module name.
func (m *HttpUpload) Name() string { return ModuleName }

// StreamFeature returns module stream feature element.
func (m *HttpUpload) StreamFeature(ctx context.Context, domain string) (stravaganza.Element, error) {
	return nil, nil
}

// func (m *HttpUpload) onElementRecv(execCtx *hook.ExecutionContext) error {

// }
