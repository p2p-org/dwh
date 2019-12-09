package imgresizer_test

import (
	"testing"

	"github.com/corestario/dwh/x/imgresizer"
	"github.com/stretchr/testify/assert"
)

var testImages = []string{
	`{"owner":"user1","token_id":"TOKEN_1","url":"http://www.redwoodsoft.com/dru/museum/gfx/gff/sample/images/png/marble24.png"}`,
	`{"owner":"user2","token_id":"TOKEN_2","url":"https://upload.wikimedia.org/wikipedia/commons/7/78/FusionintheSun.svg"}`,
	`{"owner":"user3","token_id":"TOKEN_3","url":"https://upload.wikimedia.org/wikipedia/commons/c/c5/JPEG_example_down.jpg"}`,
	`{"owner":"user4","token_id":"TOKEN_4","url":"https://www.fileformat.info/format/tiff/sample/c44cf1326c2240d38e9fca073bd7a805/download"}`,
	`{"owner":"user5","token_id":"TOKEN_5","url":"https://i.pinimg.com/originals/d5/44/ff/d544ffca4ecb461fc19da7e384cbc6d5.gif"}`,
	`{"owner":"user6","token_id":"TOKEN_6","url":"https://www.fileformat.info/format/bmp/sample/7223b4e69ae34afc8981bc11a6bb7e40/download"}`,
	`{"owner":"user7","token_id":"TOKEN_7","url":"https://res.cloudinary.com/ireaderinokun/image/upload/v1542636766/bitsofcode/comparison.webp"}`,
}

func TestRun(t *testing.T) {
	w, err := imgresizer.NewImageProcessingWorker("", "")
	assert.Nil(t, err)
	for _, v := range testImages {
		err := w.ProcessMessage([]byte(v))
		assert.Nil(t, err)
	}
}
