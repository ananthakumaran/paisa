package server

import (
	"math"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Node struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Link struct {
	Source uint    `json:"source"`
	Target uint    `json:"target"`
	Value  float64 `json:"value"`
}

type Pair struct {
	Source uint `json:"source"`
	Target uint `json:"target"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

func GetCurrentExpense(db *gorm.DB) map[string][]posting.Posting {
	expenses := query.Init(db).Unbudgeted().LastNMonths(3).Like("Expenses:%").NotLike("Expenses:Tax").All()
	return utils.GroupByMonth(expenses)
}

func GetExpense(db *gorm.DB) gin.H {
	expenses := query.Init(db).Unbudgeted().Like("Expenses:%").NotLike("Expenses:Tax").All()
	incomes := query.Init(db).Unbudgeted().Like("Income:%").All()
	investments := query.Init(db).Unbudgeted().Like("Assets:%").NotLike("Assets:Checking").All()
	taxes := query.Init(db).Unbudgeted().Like("Expenses:Tax").All()
	postings := query.Init(db).Unbudgeted().All()

	graph := make(map[string]Graph)
	for fy, ps := range utils.GroupByFY(postings) {
		graph[fy] = computeGraph(ps)
	}

	return gin.H{
		"expenses": expenses,
		"month_wise": gin.H{
			"expenses":    utils.GroupByMonth(expenses),
			"incomes":     utils.GroupByMonth(incomes),
			"investments": utils.GroupByMonth(investments),
			"taxes":       utils.GroupByMonth(taxes)},
		"year_wise": gin.H{
			"expenses":    utils.GroupByFY(expenses),
			"incomes":     utils.GroupByFY(incomes),
			"investments": utils.GroupByFY(investments),
			"taxes":       utils.GroupByFY(taxes)},
		"graph": graph}
}

func computeGraph(postings []posting.Posting) Graph {
	nodes := make(map[string]Node)
	links := make(map[Pair]float64)

	var nodeID uint = 0

	transactions := transaction.Build(postings)

	for _, p := range postings {
		_, ok := nodes[p.Account]
		if !ok {
			nodeID++
			nodes[p.Account] = Node{ID: nodeID, Name: p.Account}
		}

	}

	for _, t := range transactions {
		from := lo.Filter(t.Postings, func(p posting.Posting, _ int) bool { return p.Amount < 0 })
		to := lo.Filter(t.Postings, func(p posting.Posting, _ int) bool { return p.Amount > 0 })

		for _, f := range from {
			for math.Abs(f.Amount) > 0.1 && len(to) > 0 {
				top := to[0]
				if top.Amount > -f.Amount {
					links[Pair{Source: nodes[f.Account].ID, Target: nodes[top.Account].ID}] += -f.Amount
					top.Amount -= f.Amount
					f.Amount = 0
				} else {
					links[Pair{Source: nodes[f.Account].ID, Target: nodes[top.Account].ID}] += top.Amount
					f.Amount += top.Amount
					to = to[1:]
				}
			}
		}
	}

	return Graph{Nodes: lo.Values(nodes), Links: lo.Map(lo.Keys(links), func(k Pair, _ int) Link {
		return Link{Source: k.Source, Target: k.Target, Value: links[k]}
	})}

}
