package characterservice

import "errors"


var(
	ErrSkinIsNotOpened = errors.New("skin is not opened")
	ErrSkinIsNotExist = errors.New("skin is not exist")
)