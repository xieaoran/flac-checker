package models

type FileMeta struct {
	FilePath string

	Title      string
	Artist     string
	Album      string
	SampleRate int
	BitDepth   int

	AudioMD5 string
}
