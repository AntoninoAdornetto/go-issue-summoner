package lexer

type Comment struct {
	Title          []byte
	Description    []byte
	TokenIndex     int
	Source         []byte
	SourceFileName string
}

func (c *Comment) Prepare(fileName string, index int) {
	c.TokenIndex = index
	c.SourceFileName = fileName
}

func (c *Comment) Validate() bool {
	return len(c.Source) > 0
}
