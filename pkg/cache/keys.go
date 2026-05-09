package cache

import "fmt"

// Key patterns for different cache types
const (
	// Post cache keys
	KeyPostByID    = "post:id:%d"      // post:id:123
	KeyPostBySlug  = "post:slug:%s"    // post:slug:my-post-slug
	KeyPostList    = "post:list:%s"    // post:list:page:1:category:2
	KeyPostCount   = "post:count:%s"   // post:count:category:2

	// Taxonomy cache keys
	KeyTermByID      = "term:id:%d"        // term:id:123
	KeyTermBySlug    = "term:slug:%s"      // term:slug:my-category
	KeyTermList      = "term:list:%s"      // term:list:category:page:1
	KeyTermAllOfType = "term:all:%s"       // term:all:category
	KeyTermAllTags   = "term:all:tag"

	// User cache keys
	KeyUserByID       = "user:id:%d"       // user:id:123
	KeyUserByUsername = "user:username:%s" // user:username:john

	// Theme cache keys
	KeyThemeInfo      = "theme:info:%s"    // theme:info:default
	KeyThemeTemplates = "theme:templates:%s" // theme:templates:default

	// Cache TTLs
	TTLShort  = 300  // 5 minutes
	TTLMedium = 900  // 15 minutes
	TTLLong   = 3600 // 1 hour
	TTLDay    = 86400 // 24 hours
)

// PostKey generates a post cache key by ID
func PostKey(id uint) string {
	return fmt.Sprintf(KeyPostByID, id)
}

// PostSlugKey generates a post cache key by slug
func PostSlugKey(slug string) string {
	return fmt.Sprintf(KeyPostBySlug, slug)
}

// PostListKey generates a post list cache key
func PostListKey(page, pageSize int, categoryID uint, status string) string {
	return fmt.Sprintf("post:list:page:%d:size:%d:cat:%d:status:%s", page, pageSize, categoryID, status)
}

// TermKey generates a term cache key by ID
func TermKey(id uint) string {
	return fmt.Sprintf(KeyTermByID, id)
}

// TermSlugKey generates a term cache key by slug
func TermSlugKey(slug string) string {
	return fmt.Sprintf(KeyTermBySlug, slug)
}

// TermListKey generates a term list cache key
func TermListKey(taxonomy string, page, pageSize int) string {
	return fmt.Sprintf(KeyTermList, taxonomy+":page:"+fmt.Sprintf("%d", page))
}

// TermAllOfTypeKey generates a key for all terms of a specific taxonomy
func TermAllOfTypeKey(taxonomy string) string {
	return fmt.Sprintf(KeyTermAllOfType, taxonomy)
}

// UserKey generates a user cache key by ID
func UserKey(id uint) string {
	return fmt.Sprintf(KeyUserByID, id)
}

// UserUsernameKey generates a user cache key by username
func UserUsernameKey(username string) string {
	return fmt.Sprintf(KeyUserByUsername, username)
}

// ThemeKey generates a theme cache key
func ThemeKey(name string) string {
	return fmt.Sprintf(KeyThemeInfo, name)
}
