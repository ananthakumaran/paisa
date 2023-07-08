package query

import (
	"github.com/ananthakumaran/paisa/internal/model/commodity"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Query struct {
	context *gorm.DB
	order   string
}

func Init(db *gorm.DB) *Query {
	return &Query{context: db, order: "ASC"}
}

func (q *Query) Desc() *Query {
	q.order = "DESC"
	return q
}

func (q *Query) Commodities(commodities []commodity.Commodity) *Query {
	q.context = q.context.Where("commodity in ?", lo.Map(commodities, func(c commodity.Commodity, _ int) string { return c.Name }))
	return q
}

func (q *Query) Credit() *Query {
	q.context = q.context.Where("amount > 0")
	return q
}

func (q *Query) AccountPrefix(account string) *Query {
	q.context = q.context.Where("account like ? or account = ?", account+":%", account)
	return q
}

func (q *Query) Like(account string) *Query {
	q.context = q.context.Where("account like ?", account)
	return q
}

func (q *Query) OrLike(account string) *Query {
	q.context = q.context.Or("account like ?", account)
	return q
}

func (q *Query) NotLike(account string) *Query {
	q.context = q.context.Where("account not like ?", account)
	return q
}

func (q *Query) Or(query interface{}, args ...interface{}) *Query {
	q.context = q.context.Or(query, args...)
	return q
}

func (q *Query) Where(query interface{}, args ...interface{}) *Query {
	q.context = q.context.Where(query, args...)
	return q
}

func (q *Query) All() []posting.Posting {
	var postings []posting.Posting
	result := q.context.Order("date " + q.order + ", amount desc").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return postings
}

func (q *Query) First() posting.Posting {
	var posting posting.Posting
	result := q.context.Order("date " + q.order + ", amount desc").First(&posting)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return posting
}
