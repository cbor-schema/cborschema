// (c) 2022-2022, CBOR Schema Group. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNode(t *testing.T) {
	assert := assert.New(t)

	n := &Node{}
	assert.Equal(InvalidType, n.SimpleType())
	assert.Error(n.Valid())
}
