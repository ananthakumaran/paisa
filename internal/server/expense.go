package server

import (
	"sort"
	"strings"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/transaction"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Node struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Link struct {
	Source uint            `json:"source"`
	Target uint            `json:"target"`
	Value  decimal.Decimal `json:"value"`
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
	expenses := query.Init(db).LastNMonths(3).Like("Expenses:%").NotAccountPrefix("Expenses:Tax").All()
	return utils.GroupByMonth(expenses)
}

func GetExpense(db *gorm.DB) gin.H {
	expenses := query.Init(db).Like("Expenses:%").NotAccountPrefix("Expenses:Tax").All()
	incomes := query.Init(db).Like("Income:%").All()
	investments := query.Init(db).Like("Assets:%").NotAccountPrefix("Assets:Checking").All()
	taxes := query.Init(db).AccountPrefix("Expenses:Tax").All()
	postings := query.Init(db).All()

	graph := make(map[string]Graph)
	for fy, ps := range utils.GroupByFY(postings) {
		graph[fy] = sortGraph(computeHierarchyGraph(ps))
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

func sortGraph(graph Graph) Graph {
	nodes := graph.Nodes
	sort.Slice(nodes, func(i, j int) bool {
		return graph.Nodes[i].Name < graph.Nodes[j].Name
	})

	links := graph.Links
	sort.Slice(links, func(i, j int) bool {
		return graph.Links[i].Source < graph.Links[j].Source || (graph.Links[i].Source == graph.Links[j].Source && graph.Links[i].Target < graph.Links[j].Target)
	})
	return Graph{
		Nodes: nodes,
		Links: links,
	}

}

func computeHierarchyGraph(postings []posting.Posting) Graph {
	nodes := make(map[string]Node)
	links := make(map[Pair]decimal.Decimal)

	var nodeID uint = 0

	transactions := transaction.Build(postings)

	for _, p := range postings {
		addNode(&nodeID, &nodes, p.Account)
	}

	for _, t := range transactions {
		from := lo.Filter(t.Postings, func(p posting.Posting, _ int) bool { return p.Amount.LessThan(decimal.Zero) })
		to := lo.Filter(t.Postings, func(p posting.Posting, _ int) bool { return p.Amount.GreaterThan(decimal.Zero) })

		for _, f := range from {
			for f.Amount.Abs().GreaterThan(decimal.NewFromFloat(0.1)) && len(to) > 0 {
				top := to[0]
				if top.Amount.GreaterThan(f.Amount.Neg()) {
					addLink(f.Account, top.Account, f.Amount.Neg(), &nodes, &links)
					top.Amount = top.Amount.Sub(f.Amount)
					f.Amount = decimal.Zero
				} else {
					addLink(f.Account, top.Account, top.Amount, &nodes, &links)
					f.Amount = f.Amount.Add(top.Amount)
					to = to[1:]
				}
			}
		}
	}

	return Graph{Nodes: lo.Values(nodes), Links: lo.Map(lo.Keys(links), func(k Pair, _ int) Link {
		return Link{Source: k.Source, Target: k.Target, Value: links[k]}
	})}

}

func addNode(nodeID *uint, nodes *map[string]Node, account string) {
	if account == "" {
		return
	}

	_, ok := (*nodes)[account]
	if !ok {
		if strings.HasPrefix(account, "Income:") || strings.HasPrefix(account, "Expenses:") {
			parts := strings.Split(account, ":")
			addNode(nodeID, nodes, strings.Join(parts[:len(parts)-1], ":"))

		}

		(*nodeID)++
		(*nodes)[account] = Node{ID: *nodeID, Name: account}
	}
}

func addLink(source string, target string, amount decimal.Decimal, nodes *map[string]Node, links *map[Pair]decimal.Decimal) {

	sparts := strings.Split(source, ":")
	if sparts[0] == "Income" {
		for len(sparts) > 1 {
			s := strings.Join(sparts, ":")
			t := strings.Join(sparts[:len(sparts)-1], ":")
			(*links)[Pair{Source: (*nodes)[s].ID, Target: (*nodes)[t].ID}] = (*links)[Pair{Source: (*nodes)[s].ID, Target: (*nodes)[t].ID}].Add(amount)
			sparts = sparts[:len(sparts)-1]
		}
		source = strings.Join(sparts, ":")

	}

	tparts := strings.Split(target, ":")
	if tparts[0] == "Expenses" {
		for len(tparts) > 1 {
			t := strings.Join(tparts, ":")
			s := strings.Join(tparts[:len(tparts)-1], ":")
			(*links)[Pair{Source: (*nodes)[s].ID, Target: (*nodes)[t].ID}] = (*links)[Pair{Source: (*nodes)[s].ID, Target: (*nodes)[t].ID}].Add(amount)
			tparts = tparts[:len(tparts)-1]
		}
		target = strings.Join(tparts, ":")
	}

	(*links)[Pair{Source: (*nodes)[source].ID, Target: (*nodes)[target].ID}] = (*links)[Pair{Source: (*nodes)[source].ID, Target: (*nodes)[target].ID}].Add(amount)
}
