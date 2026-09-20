// Package websiteidentity holds the public face of goholic.in.
//
// Think of this package as the nameplate on the front door.
// Every page reads the same name, author, and home city from here.
package websiteidentity

// WebsiteIdentity is the nameplate for the whole blog.
type WebsiteIdentity struct {
	WebsiteDisplayName string
	WebsiteHomeAddress string
	AuthorFullName     string
	AuthorJobTitle     string
	AuthorHomeCity     string
	AuthorEmailAddress string
	AuthorGithubPage   string
	AuthorLinkedinPage string
	ShortWelcomeLine   string
}

// NewGoholicWebsiteIdentity returns the real nameplate for goholic.in.
func NewGoholicWebsiteIdentity() WebsiteIdentity {
	return WebsiteIdentity{
		WebsiteDisplayName: "Goholic",
		WebsiteHomeAddress: "https://goholic.in",
		AuthorFullName:     "Abir Sarkar",
		AuthorJobTitle:     "Software Engineer II · Backend (Go)",
		AuthorHomeCity:     "Kolkata, India",
		AuthorEmailAddress: "webdev.abir@gmail.com",
		AuthorGithubPage:   "https://github.com/anti-gravity-bit",
		AuthorLinkedinPage: "https://linkedin.com/in/abir-sarkar-dev",
		ShortWelcomeLine:   "Notes from a Go backend engineer who likes clean services, honest latency, and small working programs.",
	}
}
