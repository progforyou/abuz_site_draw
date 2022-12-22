package axcrudobject

type Paginator struct {
	Pages     []int64
	Page      int64
	NextPage  int64
	LastPage  int64
	PrevPage  int64
	FirstPage int64
	Opts      map[string]interface{}
}

func NewPaginator(page int64, pagesCount int64) Paginator {
	if pagesCount == 0 {
		return Paginator{Pages: []int64{}, Page: 1, FirstPage: 1, LastPage: 1, Opts: map[string]interface{}{}}
	}
	res := Paginator{Pages: []int64{}, Page: page, FirstPage: 1, LastPage: pagesCount}
	if page < pagesCount {
		res.NextPage = page + 1
	}
	if page > 1 {
		res.PrevPage = page - 1
	}

	if pagesCount < 11 {
		for i := int64(0); i < pagesCount; i++ {
			res.Pages = append(res.Pages, i+1)
		}
		return res
	}

	if page < 6 {
		for i := int64(0); i < 8; i++ {
			res.Pages = append(res.Pages, i+1)
		}
		res.Pages = append(res.Pages, 0)
		for i := pagesCount - 3; i < pagesCount; i++ {
			res.Pages = append(res.Pages, i+1)
		}
	} else if page > pagesCount-6 {
		for i := int64(0); i < 3; i++ {
			res.Pages = append(res.Pages, i+1)
		}
		res.Pages = append(res.Pages, 0)
		for i := pagesCount - 7; i < pagesCount; i++ {
			res.Pages = append(res.Pages, i+1)
		}
	} else {
		for i := int64(0); i < 3; i++ {
			res.Pages = append(res.Pages, i+1)
		}
		res.Pages = append(res.Pages, 0)
		for i := page - 3; i < page+2; i++ {
			res.Pages = append(res.Pages, i+1)
		}
		res.Pages = append(res.Pages, 0)
		for i := pagesCount - 3; i < pagesCount; i++ {
			res.Pages = append(res.Pages, i+1)
		}
	}
	return res
}
