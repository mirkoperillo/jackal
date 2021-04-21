// Copyright 2021 The jackal Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xep0012

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackal-xmpp/sonar"
	"github.com/jackal-xmpp/stravaganza/v2"
	"github.com/jackal-xmpp/stravaganza/v2/jid"
	"github.com/ortuman/jackal/event"
	coremodel "github.com/ortuman/jackal/model/core"
	lastmodel "github.com/ortuman/jackal/model/last"
	rostermodel "github.com/ortuman/jackal/model/roster"
	"github.com/ortuman/jackal/module"
	xmpputil "github.com/ortuman/jackal/util/xmpp"
	"github.com/stretchr/testify/require"
)

func TestLast_GetServerUptime(t *testing.T) {
	// given
	routerMock := &routerMock{}

	var respStanzas []stravaganza.Stanza
	routerMock.RouteFunc = func(ctx context.Context, stanza stravaganza.Stanza) ([]jid.JID, error) {
		respStanzas = append(respStanzas, stanza)
		return nil, nil
	}
	m := &Last{
		router: routerMock,
		sn:     sonar.New(),
	}

	// when
	_ = m.Start(context.Background())
	defer func() { _ = m.Stop(context.Background()) }()

	time.Sleep(time.Second * 2) // 2 seconds uptime

	iq, _ := stravaganza.NewIQBuilder().
		WithAttribute(stravaganza.ID, uuid.New().String()).
		WithAttribute(stravaganza.Type, stravaganza.GetType).
		WithAttribute(stravaganza.From, "ortuman@jackal.im/chamber").
		WithAttribute(stravaganza.To, "jackal.im").
		WithChild(
			stravaganza.NewBuilder("query").
				WithAttribute(stravaganza.Namespace, lastActivityNamespace).
				Build(),
		).
		BuildIQ()
	_ = m.ProcessIQ(context.Background(), iq)

	// then
	require.Len(t, respStanzas, 1)

	require.Equal(t, "iq", respStanzas[0].Name())
	require.Equal(t, stravaganza.ResultType, respStanzas[0].Attribute(stravaganza.Type))

	q := respStanzas[0].ChildNamespace("query", lastActivityNamespace)
	require.NotNil(t, q)
	require.True(t, len(q.Attribute("seconds")) > 0)
	require.NotEqual(t, "0", q.Attribute("seconds"))
}

func TestLast_GetAccountLastActivityOnline(t *testing.T) {
	// given
	routerMock := &routerMock{}

	var respStanzas []stravaganza.Stanza
	routerMock.RouteFunc = func(ctx context.Context, stanza stravaganza.Stanza) ([]jid.JID, error) {
		respStanzas = append(respStanzas, stanza)
		return nil, nil
	}
	repMock := &repositoryMock{}
	repMock.FetchRosterItemFunc = func(ctx context.Context, username string, jid string) (*rostermodel.Item, error) {
		switch username {
		case "noelia":
			return &rostermodel.Item{Username: "noelia", JID: "ortuman@jackal.im", Subscription: rostermodel.From}, nil
		case "romeo":
			return &rostermodel.Item{Username: "romeo", JID: "ortuman@jackal.im", Subscription: rostermodel.Both}, nil
		}
		return nil, nil
	}
	repMock.FetchLastFunc = func(ctx context.Context, username string) (*lastmodel.Last, error) {
		return &lastmodel.Last{
			Username: "noelia",
			Seconds:  time.Now().Unix() - 100,
			Status:   "Heading home",
		}, nil
	}

	hMock := &hostsMock{}
	hMock.IsLocalHostFunc = func(h string) bool { return h == "jackal.im" }

	resMngMock := &resourceManagerMock{}

	jd0, _ := jid.NewWithString("noelia@jackal.im/yard", true)
	resMngMock.GetResourcesFunc = func(ctx context.Context, username string) ([]coremodel.Resource, error) {
		if username != "noelia" {
			return nil, nil
		}
		return []coremodel.Resource{
			{JID: jd0},
		}, nil
	}

	m := &Last{
		router: routerMock,
		rep:    repMock,
		hosts:  hMock,
		resMng: resMngMock,
		sn:     sonar.New(),
	}

	// when
	iq1, _ := stravaganza.NewIQBuilder().
		WithAttribute(stravaganza.ID, uuid.New().String()).
		WithAttribute(stravaganza.Type, stravaganza.GetType).
		WithAttribute(stravaganza.From, "ortuman@jackal.im/chamber").
		WithAttribute(stravaganza.To, "noelia@jackal.im").
		WithChild(
			stravaganza.NewBuilder("query").
				WithAttribute(stravaganza.Namespace, lastActivityNamespace).
				Build(),
		).
		BuildIQ()
	iq2, _ := stravaganza.NewIQBuilder().
		WithAttribute(stravaganza.ID, uuid.New().String()).
		WithAttribute(stravaganza.Type, stravaganza.GetType).
		WithAttribute(stravaganza.From, "ortuman@jackal.im/chamber").
		WithAttribute(stravaganza.To, "romeo@jackal.im").
		WithChild(
			stravaganza.NewBuilder("query").
				WithAttribute(stravaganza.Namespace, lastActivityNamespace).
				Build(),
		).
		BuildIQ()

	_ = m.ProcessIQ(context.Background(), iq1)
	_ = m.ProcessIQ(context.Background(), iq2)

	// then
	require.Len(t, respStanzas, 2)

	require.Equal(t, "iq", respStanzas[0].Name())
	require.Equal(t, stravaganza.ResultType, respStanzas[0].Attribute(stravaganza.Type))

	q1 := respStanzas[0].ChildNamespace("query", lastActivityNamespace)
	require.NotNil(t, q1)
	require.True(t, len(q1.Attribute("seconds")) > 0)
	require.Equal(t, "0", q1.Attribute("seconds"))

	require.Equal(t, "iq", respStanzas[1].Name())
	require.Equal(t, stravaganza.ResultType, respStanzas[1].Attribute(stravaganza.Type))

	q2 := respStanzas[1].ChildNamespace("query", lastActivityNamespace)
	require.NotNil(t, q2)
	require.True(t, len(q2.Attribute("seconds")) > 0)
	require.NotEqual(t, "0", q2.Attribute("seconds"))
}

func TestLast_Forbidden(t *testing.T) {
	// given
	routerMock := &routerMock{}

	var respStanzas []stravaganza.Stanza
	routerMock.RouteFunc = func(ctx context.Context, stanza stravaganza.Stanza) ([]jid.JID, error) {
		respStanzas = append(respStanzas, stanza)
		return nil, nil
	}
	repMock := &repositoryMock{}
	repMock.FetchRosterItemFunc = func(ctx context.Context, username string, jid string) (*rostermodel.Item, error) {
		return nil, nil
	}
	hMock := &hostsMock{}
	hMock.IsLocalHostFunc = func(h string) bool { return h == "jackal.im" }

	m := &Last{
		router: routerMock,
		rep:    repMock,
		hosts:  hMock,
		sn:     sonar.New(),
	}

	// when
	iq, _ := stravaganza.NewIQBuilder().
		WithAttribute(stravaganza.ID, uuid.New().String()).
		WithAttribute(stravaganza.Type, stravaganza.GetType).
		WithAttribute(stravaganza.From, "ortuman@jackal.im/chamber").
		WithAttribute(stravaganza.To, "noelia@jackal.im").
		WithChild(
			stravaganza.NewBuilder("query").
				WithAttribute(stravaganza.Namespace, lastActivityNamespace).
				Build(),
		).
		BuildIQ()

	_ = m.ProcessIQ(context.Background(), iq)

	// then
	require.Len(t, respStanzas, 1)

	require.Equal(t, "iq", respStanzas[0].Name())
	require.Equal(t, stravaganza.ErrorType, respStanzas[0].Attribute(stravaganza.Type))

	errEl := respStanzas[0].Child("error")
	require.NotNil(t, errEl)
	require.NotNil(t, errEl.ChildNamespace("forbidden", "urn:ietf:params:xml:ns:xmpp-stanzas"))
}

func TestLast_InterceptStanza(t *testing.T) {
	// given
	routerMock := &routerMock{}

	var respStanzas []stravaganza.Stanza
	routerMock.RouteFunc = func(ctx context.Context, stanza stravaganza.Stanza) ([]jid.JID, error) {
		respStanzas = append(respStanzas, stanza)
		return nil, nil
	}
	repMock := &repositoryMock{}
	repMock.FetchRosterItemFunc = func(ctx context.Context, username string, jid string) (*rostermodel.Item, error) {
		return nil, nil
	}
	hMock := &hostsMock{}
	hMock.IsLocalHostFunc = func(h string) bool { return h == "jackal.im" }

	m := &Last{
		router: routerMock,
		rep:    repMock,
		hosts:  hMock,
		sn:     sonar.New(),
	}
	// when
	iq, _ := stravaganza.NewIQBuilder().
		WithAttribute(stravaganza.ID, uuid.New().String()).
		WithAttribute(stravaganza.Type, stravaganza.GetType).
		WithAttribute(stravaganza.From, "ortuman@jackal.im/chamber").
		WithAttribute(stravaganza.To, "noelia@jackal.im/yard").
		WithChild(
			stravaganza.NewBuilder("query").
				WithAttribute(stravaganza.Namespace, lastActivityNamespace).
				Build(),
		).
		BuildIQ()

	_, err := m.InterceptStanza(context.Background(), iq, 0)

	// then
	require.NotNil(t, err)
	require.Equal(t, module.ErrInterceptStanzaInterrupted, err)

	require.Len(t, respStanzas, 1)

	require.Equal(t, "iq", respStanzas[0].Name())
	require.Equal(t, "noelia@jackal.im/yard", respStanzas[0].Attribute(stravaganza.From))
	require.Equal(t, stravaganza.ErrorType, respStanzas[0].Attribute(stravaganza.Type))

	errEl := respStanzas[0].Child("error")
	require.NotNil(t, errEl)
	require.NotNil(t, errEl.ChildNamespace("forbidden", "urn:ietf:params:xml:ns:xmpp-stanzas"))
}

func TestLast_ProcessPresence(t *testing.T) {
	// given
	rep := &repositoryMock{}
	rep.UpsertLastFunc = func(ctx context.Context, last *lastmodel.Last) error {
		return nil
	}

	sn := sonar.New()
	bl := &Last{
		rep: rep,
		sn:  sn,
	}
	// when
	_ = bl.Start(context.Background())
	defer func() { _ = bl.Stop(context.Background()) }()

	jd0, _ := jid.NewWithString("ortuman@jackal.im/yard", true)
	_ = sn.Post(context.Background(), sonar.NewEventBuilder(event.C2SStreamPresenceReceived).
		WithInfo(&event.C2SStreamEventInfo{
			JID:    jd0,
			Stanza: xmpputil.MakePresence(jd0, jd0.ToBareJID(), stravaganza.UnavailableType, nil),
		}).
		Build(),
	)

	// then
	require.Len(t, rep.UpsertLastCalls(), 1)
}