package types

type Role string
type ContextKey string
type PostStatus string //	@name	TypePostStatus

const (
	Author    Role       = "author"
	Reader    Role       = "reader"
	CtxUser   ContextKey = "user"
	Draft     PostStatus = "draft"     //	@name	DraftStatus
	Published PostStatus = "published" //	@name	PublishedStatus
)
