package pike

import "path/filepath"

// File provides a way to transmit file data through a Graph. It is an
// interface to allow users to create their own File structs for custom
// applications.
type File interface {
	Root() string
	SetRoot(root string)
	Name() string
	SetName(name string)
	Data() []byte
	SetData(data []byte)

	Copy() File
	// the fully-qualified path to the file
	Fullpath() string
	// set the file extension
	SetExt(ext string)
}

// BaseFile is the standard implementation of File
type BaseFile struct {
	root string
	name string
	data []byte
}

func (self *BaseFile) Root() string {
	return self.root
}
func (self *BaseFile) SetRoot(root string) {
	self.root = root
}
func (self *BaseFile) Name() string {
	return self.name
}
func (self *BaseFile) SetName(name string) {
	self.name = name
}
func (self *BaseFile) Data() []byte {
	return self.data
}
func (self *BaseFile) SetData(data []byte) {
	self.data = data
}

// NewFile is a constructor for BaseFile
func NewFile(root, name string, data []byte) File {
	return &BaseFile{root, name, data}
}

func (self *BaseFile) Copy() File {
	data := make([]byte, len(self.Data()))
	copy(data, self.Data())
	return &BaseFile{self.Root(), self.Name(), data}
}

func (file *BaseFile) Fullpath() string {
	if file.Root() != "" {
		return filepath.Join(file.Root(), file.Name())
	}
	return file.Name()
}

func (file *BaseFile) SetExt(ext string) {
	oldExt := filepath.Ext(file.Name())
	file.SetName(file.Name()[:len(file.Name())-len(oldExt)] + ext)
}
