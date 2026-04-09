package aggregate

import (
	"github.com/GuanceCloud/cliutils/point"
)

func packetPointCount(packet *DataPacket) int32 {
	if packet == nil {
		return 0
	}

	return packet.PointCount
}

func (packet *DataPacket) WalkRawPBPoints(fn func([]byte) bool) error {
	if packet == nil || fn == nil {
		return nil
	}

	return point.WalkPBPointsPayload(packet.PointsPayload, fn)
}
