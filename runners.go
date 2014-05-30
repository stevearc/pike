package pike

// Runnable is the piece of a Node that performs the file operations.
type Runnable interface {
	Run(in, out []chan File)
	Copy() Runnable
}

// FxnRunnable is the most simple Runnable. It is just a function.
type FxnRunnable func(in, out []chan File)

func (self FxnRunnable) Run(in, out []chan File) {
	self(in, out)
}

func (self FxnRunnable) Copy() Runnable {
	return self
}

// CacheRunnable is a function that caches file state between runs. This is
// used for the change filter nodes.
type CacheRunnable struct {
	Fxn   func(in, out []chan File, cache map[string]File)
	Cache map[string]File
}

func (self *CacheRunnable) Run(in, out []chan File) {
	self.Fxn(in, out, self.Cache)
}

func (self *CacheRunnable) Copy() Runnable {
	newMap := make(map[string]File, len(self.Cache))
	for key, val := range self.Cache {
		newMap[key] = val
	}
	return &CacheRunnable{self.Fxn, newMap}
}

// NewCacheRunnable is a constructor for CacheRunnable.
func NewCacheRunnable(run func(in, out []chan File, cache map[string]File)) Runnable {
	return &CacheRunnable{run, make(map[string]File)}
}
