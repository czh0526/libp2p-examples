package common

const (
	CODE_SUCCESS        = 0
	CODE_FILE_NON_EXIST = 1
)

type Error struct {
	Code int
	Msg  string
}
