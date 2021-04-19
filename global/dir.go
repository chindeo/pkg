package global

type DeviceType int64

const (
	DeviceTypeEmpty DeviceType = iota
	DeviceTypeNIS
	DeviceTypeBIS
	DeviceTypeWEBAPP
	DeviceTypeNWS
)

func (dy DeviceType) String() string {
	switch dy {
	case DeviceTypeNIS:
		return "nis"
	case DeviceTypeBIS:
		return "bis"
	case DeviceTypeWEBAPP:
		return "webapp"
	case DeviceTypeNWS:
		return "nws"
	case DeviceTypeEmpty:
		return ""
	default:
		return "other"
	}
}
