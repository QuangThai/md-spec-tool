package share

// StoreReader defines read operations on the share store
type StoreReader interface {
	GetShare(key string) (*Share, error)
	ListPublic() []*Share
	ListComments(key string) ([]Comment, error)
}

// StoreWriter defines write operations on the share store
type StoreWriter interface {
	CreateShare(input CreateShareInput) (*Share, error)
	UpdateShare(key string, isPublic *bool, allowComments *bool) (*Share, error)
	AddComment(key string, input CommentInput) (Comment, error)
	UpdateComment(key, commentID string, resolved bool) (Comment, error)
}

// Store is a complete interface combining read and write operations
type StoreInterface interface {
	StoreReader
	StoreWriter
}

// Ensure Store implements the interface
var _ StoreInterface = (*Store)(nil)
