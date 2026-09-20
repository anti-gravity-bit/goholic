package articlelibrarian

// PlantFirstStoriesIfTheShelfIsEmpty writes three published essays
// the first time a brand new cabinet is opened.
func (articleLibrarian *ArticleLibrarian) PlantFirstStoriesIfTheShelfIsEmpty() error {
	existingArticles, listError := articleLibrarian.ListEveryArticleForTheWriter()
	if listError != nil {
		return listError
	}

	if len(existingArticles) > 0 {
		return nil
	}

	type firstStory struct {
		publicWebAddressSlug string
		articleTitle         string
		articleSummary       string
		articleBodyMarkdown  string
	}

	firstStories := []firstStory{
		{
			publicWebAddressSlug: "packages-should-tell-the-story",
			articleTitle:         "Packages should tell the story",
			articleSummary:       "If a stranger can read the folder names and guess the architecture, the design is doing its job.",
			articleBodyMarkdown: `# Packages should tell the story

A blog is a tiny company.

- The **article** room decides what a story is.
- The **shelf** room only promises to keep stories.
- The **librarian** room follows the house rules.
- The **webpage** room talks to browsers.

If those four names are clear, you already understand Goholic.

## Why this stays small

I ship production Go services for a living. The ones that last are the ones a tired teammate can open at midnight and still explain out loud.

This site is that idea with the volume turned down.`,
		},
		{
			publicWebAddressSlug: "solid-without-the-ceremony",
			articleTitle:         "SOLID without the ceremony",
			articleSummary:       "Five grown-up rules, said like a five year old, and used to keep this blog replaceable.",
			articleBodyMarkdown: `# SOLID without the ceremony

- **S**ingle job: the passport does not store articles.
- **O**pen for new cabinets: sqlite today, postgres tomorrow, same librarian.
- **L**iskov: a memory shelf and a sqlite shelf both keep the same promise.
- **I**nterface that is small: a clock only answers the time.
- **D**epend on the promise: webpages never say the word sqlite.

That is the whole trick.`,
		},
		{
			publicWebAddressSlug: "from-kolkata-with-go",
			articleTitle:         "From Kolkata with Go",
			articleSummary:       "A short hello from Abir Sarkar, the writer behind goholic.in.",
			articleBodyMarkdown: `# From Kolkata with Go

I am Abir Sarkar, a Software Engineer II on the backend side.

Most days I work on Go services, PostgreSQL, auth, and the boring work that keeps p99 honest.

Goholic is my public notebook: small programs, clean rooms, and the occasional note about latency, tests, and WebRTC rooms.

If you want the longer version of the resume, you already know where to look. This site is for the sentences that do not fit on one page.`,
		},
	}

	for _, storyToPlant := range firstStories {
		_, writeError := articleLibrarian.WriteNewDraftArticle(
			storyToPlant.publicWebAddressSlug,
			storyToPlant.articleTitle,
			storyToPlant.articleSummary,
			storyToPlant.articleBodyMarkdown,
		)
		if writeError != nil {
			return writeError
		}

		_, publishError := articleLibrarian.ShowArticleToThePublic(storyToPlant.publicWebAddressSlug)
		if publishError != nil {
			return publishError
		}
	}

	return nil
}
