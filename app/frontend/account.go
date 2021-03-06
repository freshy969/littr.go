package frontend

import (
	"fmt"
	"github.com/mariusor/littr.go/app"
	"github.com/mariusor/littr.go/internal/errors"
	"github.com/mariusor/qstring"
	"net/http"

	"github.com/go-chi/chi"
)

// Paginator is an interface for paginating collections
type Paginator interface {
	NextPage() int
	PrevPage() int
}

type itemListingModel struct {
	Title    string
	User     *app.Account
	Items    comments
	HideText bool
	nextPage int
	prevPage int
}

func (i itemListingModel) NextPage() int {
	return i.nextPage
}

func (i itemListingModel) PrevPage() int {
	return i.prevPage
}

type sessionAccount struct {
	Hash   []byte
	Handle string
	OAuth  app.OAuth
}

// ShowAccount serves /~handler request
func (h *handler) ShowAccount(w http.ResponseWriter, r *http.Request) {
	handle := chi.URLParam(r, "handle")

	val := r.Context().Value(app.RepositoryCtxtKey)
	accountLoader, ok := val.(app.CanLoadAccounts)
	if !ok {
		h.logger.Error("could not load account repository from Context")
		return
	}
	var err error
	a, err := accountLoader.LoadAccount(app.Filters{LoadAccountsFilter: app.LoadAccountsFilter{Handle: []string{handle}}})
	if err != nil {
		h.HandleErrors(w, r, err)
		return
	}
	if !a.IsValid() {
		h.HandleErrors(w, r, errors.NotFoundf("account %q not found", handle))
		return
	}

	filter := app.Filters{
		LoadItemsFilter: app.LoadItemsFilter{
			AttributedTo: app.Hashes{a.Hash},
		},
		MaxItems: MaxContentItems,
		Page:     1,
	}
	if err := qstring.Unmarshal(r.URL.Query(), &filter); err != nil {
		h.logger.Debug("unable to load url parameters")
	}
	if m, err := loadItems(r.Context(), filter, &h.account, h.logger); err == nil {
		m.Title = fmt.Sprintf("%s submissions", genitive(a.Handle))
		m.User = &a

		if len(m.Items) >= filter.MaxItems {
			m.nextPage = filter.Page + 1
		}
		if filter.Page > 1 {
			m.prevPage = filter.Page - 1
		}

		h.RenderTemplate(r, w, "user", m)
	} else {
		h.HandleErrors(w, r, errors.NewNotValid(err, "unable to load items"))
	}
}
