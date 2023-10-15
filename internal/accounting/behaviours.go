package accounting

import (
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/samber/lo"
)

func PostingWithBehaviours(postings []posting.Posting, behaviours []string) []posting.Posting {
	return lo.Filter(postings, func(p posting.Posting, _ int) bool {
		for _, b := range behaviours {
			if p.HasBehaviour(b) {
				return true
			}
		}
		return false
	})
}
