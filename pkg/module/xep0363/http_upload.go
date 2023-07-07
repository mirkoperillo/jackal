package xep0363

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/jackal-xmpp/stravaganza"
	stanzaerror "github.com/jackal-xmpp/stravaganza/errors/stanza"
	"github.com/ortuman/jackal/pkg/cluster/resourcemanager"
	"github.com/ortuman/jackal/pkg/hook"
	httpServer "github.com/ortuman/jackal/pkg/httpserver"
	model "github.com/ortuman/jackal/pkg/model/httpupload"
	"github.com/ortuman/jackal/pkg/router"
	"github.com/ortuman/jackal/pkg/storage/repository"
	xmpputil "github.com/ortuman/jackal/pkg/util/xmpp"
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
	config Config
	rep    repository.Repository
	//hosts  hosts
	httpServer *httpServer.HttpServer
	router     router.Router
	resMng     resourcemanager.Manager
	hk         *hook.Hooks
	logger     kitlog.Logger
}

type Config struct {
	StorageFolder string `fig:"storage_folder"`
}

// New returns a new initialized BlockList instance.
func New(
	config Config,
	router router.Router,
	//hosts *host.Hosts,
	httpServer *httpServer.HttpServer,
	resMng resourcemanager.Manager,
	rep repository.Repository,
	hk *hook.Hooks,
	logger kitlog.Logger,
) *HttpUpload {
	return &HttpUpload{
		config:     config,
		rep:        rep,
		httpServer: httpServer,
		router:     router,
		//hosts:  hosts,
		resMng: resMng,
		hk:     hk,
		logger: kitlog.With(logger, "module", ModuleName, "xep", XEPNumber),
	}
}

func (m *HttpUpload) Start(ctx context.Context) error {
	level.Info(m.logger).Log("msg", "started http_upload module")
	pathFolder := m.config.StorageFolder
	_, err := os.Stat(pathFolder)
	if os.IsNotExist(err) {
		level.Info(m.logger).Log("msg", fmt.Sprintf("storage folder %s not exist, creating...", pathFolder))
		err := os.MkdirAll(pathFolder, 0700)
		if err != nil {
			level.Error(m.logger).Log("msg", fmt.Sprintf("error creating storage folder: %s", err))
		}
	} else {
		level.Info(m.logger).Log("msg", fmt.Sprintf("storage folder %s", pathFolder))
	}

	m.httpServer.Register("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			level.Warn(m.logger).Log("msg", fmt.Sprintf("unsupported method %s for /upload enpoint", r.Method))
			return
		}
		fileContent, err := io.ReadAll(r.Body)
		if err != nil {
			level.Error(m.logger).Log("msg", fmt.Sprintf("error reading /upload body: %s", err))
		}
		err = os.WriteFile(fmt.Sprintf("%s/%s", pathFolder, "myFile.txt"), fileContent, 0700)
		if err != nil {
			level.Error(m.logger).Log("msg", fmt.Sprintf("error saving file on storage: %s", err))
		}
		level.Info(m.logger).Log("msg", "handled http file upload request")
	})

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

func (m *HttpUpload) MatchesNamespace(namespace string, _ bool) bool {
	return namespace == HttpUploadNamespace
}

func (m *HttpUpload) ProcessIQ(ctx context.Context, iq *stravaganza.IQ) error {
	switch {
	case iq.IsGet() && iq.ChildNamespace("request", HttpUploadNamespace) != nil:
		return m.upload(ctx, iq)
	default:
		_, _ = m.router.Route(ctx, xmpputil.MakeErrorStanza(iq, stanzaerror.BadRequest))
		return nil
	}
}

func (m *HttpUpload) upload(ctx context.Context, iq *stravaganza.IQ) error {
	// filename, size attribute required
	var filename string
	var filesize int
	var fileContentType string // optional
	for _, attr := range iq.Child("request").AllAttributes() {
		if !xmpputil.IsNamespaceAttr(attr) {
			switch attr.Label {
			case "filename":
				filename = attr.Value
			case "size":
				if size, err := strconv.Atoi(attr.Value); err == nil {
					filesize = size
				} else {
					_, _ = m.router.Route(ctx, xmpputil.MakeErrorStanza(iq, stanzaerror.BadRequest))
				}
			case "content-type":
				fileContentType = attr.Label
			default:
				_, _ = m.router.Route(ctx, xmpputil.MakeErrorStanza(iq, stanzaerror.BadRequest))
			}
		}
	}
	if filename == "" || filesize == 0 {
		// errors
	}
	uploadSlot := model.UploadSlot{Size: filesize, Filename: filename, ContentType: fileContentType}
	err := m.rep.InsertSlot(ctx, &uploadSlot)
	if err != nil {
		level.Error(m.logger).Log("msg", fmt.Sprintf("error storing request slot metadata: %s", err))
	}
	m.sendReply(ctx, iq, uploadSlot)
	return nil

}

// <iq from='upload.montague.tld'
//     id='step_03'
//     to='romeo@montague.tld/garden'
//     type='result'>
//   <slot xmlns='urn:xmpp:http:upload:0'>
//     <put url='https://upload.montague.tld/4a771ac1-f0b2-4a4a-9700-f2a26fa2bb67/tr%C3%A8s%20cool.jpg'>
//       <header name='Authorization'>Basic Base64String==</header>
//       <header name='Cookie'>foo=bar; user=romeo</header>
//     </put>
//     <get url='https://download.montague.tld/4a771ac1-f0b2-4a4a-9700-f2a26fa2bb67/tr%C3%A8s%20cool.jpg' />
//   </slot>

func (m *HttpUpload) sendReply(ctx context.Context, iq *stravaganza.IQ, uploadSlot model.UploadSlot) {
	putChild := stravaganza.NewBuilder("put").
		WithAttribute("url", fmt.Sprintf("https://localhost:6060/upload/%s", uploadSlot.Id)).
		Build()
	getChild := stravaganza.NewBuilder("get").
		WithAttribute("url", fmt.Sprintf("https://localhost:6060/download/%s", uploadSlot.Id)).
		Build()
	resIQ := xmpputil.MakeResultIQ(iq, stravaganza.NewBuilder("slot").
		WithAttribute(stravaganza.Namespace, HttpUploadNamespace).
		WithChildren(putChild, getChild).
		Build(),
	)
	_, _ = m.router.Route(ctx, resIQ)
}

// func (m *HttpUpload) onElementRecv(execCtx *hook.ExecutionContext) error {

// }
