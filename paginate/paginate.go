package paginate

import (
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/andrew-d/goscrape"
)

type bySelectorPaginator struct {
	sel  string
	attr string
}

// BySelector returns a Paginator that extracts the next page from a document by
// querying a given CSS selector and extracting the given HTML attribute from the
// resulting element.
func BySelector(sel, attr string) scrape.Paginator {
	return &bySelectorPaginator{
		sel: sel, attr: attr,
	}
}

func (p *bySelectorPaginator) NextPage(uri string, doc *goquery.Selection) (string, error) {
	val, found := doc.Find(p.sel).Attr(p.attr)
	if !found {
		return "", nil
	}

	// Make the URL absolute.
	base, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	attrUrl, err := url.Parse(val)
	if err != nil {
		return "", err
	}

	newUrl := base.ResolveReference(attrUrl)
	return newUrl.String(), nil
}

type byQueryParamPaginator struct {
	param string
}

// ByQueryParam returns a Paginator that returns the next page from a document by
// incrementing a given query parameter.  Note that this will paginate
// infinitely - you probably want to use the LimitPages wrapper to determine how
// many pages to paginate
func ByQueryParam(param string) scrape.Paginator {
	return &byQueryParamPaginator{param}
}

func (p *byQueryParamPaginator) NextPage(u string, _ *goquery.Selection) (string, error) {
	// Parse
	uri, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	// Parse query
	vals, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return "", err
	}

	// Find query param and increment.  If it doesn't exist, then we just stop.
	params, ok := vals[p.param]
	if !ok || len(params) < 1 {
		return "", nil
	}

	parsed, err := strconv.ParseUint(params[0], 10, 64)
	if err != nil {
		// TODO: should this be fatal?
		return "", nil
	}

	// Put everything back together
	params[0] = strconv.FormatUint(parsed+1, 10)
	vals[p.param] = params
	query := vals.Encode()
	uri.RawQuery = query
	return uri.String(), nil
}

type limitPagesPaginator struct {
	current    int
	limit      int
	underlying scrape.Paginator
}

// LimitPages is a wrapper that takes an existing Paginator and limits the number of
// pages that it returns.  This is useful when paginating infinite-scroll sites
// and the like.
//
// Note: the first page of a scrape is always retrieved and paginated, so this
// limit is actually limiting the number of *additional* pages that are paginated.
// For example, if `limit` is 1, then the initial URL provided, plus one additional
// page will be used.  Use a `limit` of 0 to prevent any pagination from occuring.
func LimitPages(limit int, underlying scrape.Paginator) scrape.Paginator {
	return &limitPagesPaginator{
		current:    0,
		limit:      limit,
		underlying: underlying,
	}
}

func (p *limitPagesPaginator) NextPage(url string, sel *goquery.Selection) (string, error) {
	// If we've already paginated this number of pages, then we stop.
	if p.current >= p.limit {
		return "", nil
	}

	p.current += 1
	return p.underlying.NextPage(url, sel)
}