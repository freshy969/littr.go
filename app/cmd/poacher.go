package cmd

import (
	"fmt"
	"github.com/juju/errors"
	"github.com/mariusor/littr.go/app"
	"github.com/mariusor/littr.go/app/db"
	"github.com/mariusor/littr.go/app/log"
	"github.com/mmcdole/gofeed"
	"net/url"
	"time"
)

var sys = app.Account{
	Hash:   "dc6f5f5bf55bc1073715c98c69fa7ca8",
	Handle: "system",
}

func PoachFeed(u string, since time.Duration) error {
	var err error
	fp := gofeed.NewParser()
	doc, err := fp.ParseURL(u)
	if err != nil {
		return errors.Annotatef(err, "failed to fetch rss feed %s", u)
	}

	baseURL := ""
	feedURL, _ := url.ParseRequestURI(u)
	if doc.Link != "" {
		feedURL, err = url.ParseRequestURI(doc.Link)
	}
	if err == nil {
		baseURL = feedURL.Host
	}
	for _, l := range doc.Items {
		acct := sys
		if l.Author.Name != "" {
			acct = app.Account{}
			if baseURL != "" {
				acct.Handle = l.Author.Name
				// @TODO(marius): this needs to have different logic based on
				//        feed source
				acct.Metadata = &app.AccountMetadata{
					WebFinger: fmt.Sprintf("%s@%s", l.Author.Name, feedURL.Host),
					URL:       fmt.Sprintf("%s://%s/~%s", feedURL.Scheme, feedURL.Host, l.Author.Name),
				}
			}
			if l.Author.Email != "" {
				acct.Email = l.Author.Email
			}
			if existing, err := db.Config.LoadAccount(app.LoadAccountsFilter{
				Handle: []string{acct.Handle},
			}); err != nil {
				acct.CreatedAt = time.Now()
				acct.UpdatedAt = acct.CreatedAt
				acct, err = db.Config.SaveAccount(acct)
				if err != nil {
					Logger.WithContext(log.Ctx{
						"key":    acct.Hash.String(),
						"handle": acct.Handle,
						"err":    err.Error(),
					}).Error("unable to save new account")
				}
				if len(acct.Handle) == 0 {
					Logger.WithContext(log.Ctx{
						"key":    acct.Hash.String(),
						"handle": acct.Handle,
					}).Error("unable to save new account")
				}
			} else {
				acct = existing
			}
			if err != nil || acct.Handle == "" {
				acct = sys
			}
		}
		item := app.Item{
			Data:        l.Link,
			SubmittedAt: *l.PublishedParsed,
			SubmittedBy: &acct,
			Title:       l.Title,
			MimeType:    app.MimeTypeURL,
		}
		item, err = db.Config.SaveItem(item)
		if err != nil {
			Logger.WithContext(log.Ctx{
				"title": item.Title,
				"data":  item.Data,
				"err":   err.Error(),
			}).Errorf("unable to save new item")
		}
		v := app.Vote{
			SubmittedBy: &acct,
			Item:        &item,
			Weight:      1 * app.ScoreMultiplier,
		}
		if _, err := db.Config.SaveVote(v); err != nil {
			Logger.WithContext(log.Ctx{
				"hash":   v.Item.Hash,
				"author": v.SubmittedBy.Handle,
				"weight": v.Weight,
			}).Error(err.Error())
		}
	}
	return nil
}