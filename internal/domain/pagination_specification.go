package domain

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PaginationSpecification struct{}

func NewPaginationSpecification() *PaginationSpecification {
	return &PaginationSpecification{}
}

func (p *PaginationSpecification) NormalizePagination(req *SearchRequest) {
	if p.isInvalidPage(req.Page) {
		req.Page = DefaultPage
	}
	if p.isInvalidPageSize(req.PageSize) {
		req.PageSize = DefaultPageSize
	}
	if p.exceedsMaxPageSize(req.PageSize) {
		req.PageSize = MaxPageSize
	}
}

func (p *PaginationSpecification) isInvalidPage(page int) bool {
	return page <= 0
}

func (p *PaginationSpecification) isInvalidPageSize(pageSize int) bool {
	return pageSize <= 0
}

func (p *PaginationSpecification) exceedsMaxPageSize(pageSize int) bool {
	return pageSize > MaxPageSize
}
