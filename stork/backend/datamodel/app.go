package datamodel

// A type defining one of the supported app types.
type AppType string

const (
	// A Kea app type.
	AppTypeKea AppType = "kea"
	// A Bind9 app type.
	AppTypeBind9 AppType = "bind9"
	// A NSD app type.
	AppTypeNSD AppType = "nsd"
)

// Converts the type to string.
func (t AppType) String() string {
	return string(t)
}

// Convenience function checking if the type is Kea.
func (t AppType) IsKea() bool {
	return t == AppTypeKea
}

// Convenience function checking if the type is BIND9.
func (t AppType) IsBind9() bool {
	return t == AppTypeBind9
}

// Convenience function checking if the type is NSD.
func (t AppType) IsNSD() bool {
	return t == AppTypeNSD
}

// Convenience function checking if the type is DNS app type.
func (t AppType) IsDNS() bool {
	return t == AppTypeNSD || t == AppTypeBind9
}