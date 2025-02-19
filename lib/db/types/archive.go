package types

type Archive struct {
	Archive bool `gorm:"column:archive" json:"archive"`
}

func (v *Archive) SetArchive(archive bool) {
	v.Archive = archive
}
