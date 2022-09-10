package query

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
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

func (q *Query) Like(account string) *Query {
	q.context = q.context.Where("account like ?", account)
	return q
}

func (q *Query) NotLike(account string) *Query {
	q.context = q.context.Where("account not like ?", account)
	return q
}

func (q *Query) All() []posting.Posting {
	var postings []posting.Posting
	result := q.context.Order("date ASC").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return postings
}

func (q *Query) First() posting.Posting {
	var posting posting.Posting
	result := q.context.Order("date " + q.order).First(&posting)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return posting
}
