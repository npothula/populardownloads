package common

import "fmt"

// FileDownloadCount ...
type FileDownloadCount struct {
	fileURI       string
	downloadCount int64
}

// NewFileDownloadCount ...
func NewFileDownloadCount(fileURI string, downloadCount int64) *FileDownloadCount {
	fileDownloadCount := &FileDownloadCount{
		fileURI:       fileURI,
		downloadCount: downloadCount,
	}
	return fileDownloadCount
}

// MaxHeap ...
type MaxHeap struct {
	heapArray []*FileDownloadCount
	size      int
	maxsize   int
}

// NewMaxHeap ...
func NewMaxHeap(maxsize int) *MaxHeap {
	MaxHeap := &MaxHeap{
		heapArray: []*FileDownloadCount{},
		size:      0,
		maxsize:   maxsize,
	}
	return MaxHeap
}

func (m *MaxHeap) leaf(index int) bool {
	if index <= m.size && index >= (m.size/2) {
		return true
	}
	return false
}

func (m *MaxHeap) parent(index int) int {
	return (index - 1) / 2
}

func (m *MaxHeap) leftchild(index int) int {
	return 2*index + 1
}

func (m *MaxHeap) rightchild(index int) int {
	return 2*index + 2
}

// Insert ...
func (m *MaxHeap) Insert(item *FileDownloadCount) error {
	if m.size <= m.maxsize {
		m.heapArray = append(m.heapArray, item)
		m.size++
		m.upHeapify(m.size - 1)
	} else {
		if item.downloadCount > m.heapArray[m.size-1].downloadCount {
			m.heapArray[m.size-1] = item
			m.upHeapify(m.size - 1)
			return nil
		}
		if item.downloadCount > m.heapArray[m.size-2].downloadCount {
			m.heapArray[m.size-2] = item
			m.upHeapify(m.size - 2)
		}
	}
	return nil
}

func (m *MaxHeap) swap(first, second int) {
	temp := m.heapArray[first]
	m.heapArray[first] = m.heapArray[second]
	m.heapArray[second] = temp
}

func (m *MaxHeap) upHeapify(index int) {
	for m.heapArray[index].downloadCount > m.heapArray[m.parent(index)].downloadCount {
		m.swap(index, m.parent(index))
	}
}

func (m *MaxHeap) downHeapify(current int) {
	if m.leaf(current) {
		return
	}
	largest := current
	leftChildIndex := m.leftchild(current)
	rightRightIndex := m.rightchild(current)
	//If current is smallest then return
	if leftChildIndex < m.size && m.heapArray[leftChildIndex].downloadCount > m.heapArray[largest].downloadCount {
		largest = leftChildIndex
	}
	if rightRightIndex < m.size && m.heapArray[rightRightIndex].downloadCount > m.heapArray[largest].downloadCount {
		largest = rightRightIndex
	}
	if largest != current {
		m.swap(current, largest)
		m.downHeapify(largest)
	}
	return
}

// Remove ...
func (m *MaxHeap) Remove() *FileDownloadCount {
	top := m.heapArray[0]
	m.heapArray[0] = m.heapArray[m.size-1]
	m.heapArray = m.heapArray[:(m.size)-1]
	m.size--
	m.downHeapify(0)
	return top
}

func mapHeapDemo() {
	topK := 2
	maxHeap := NewMaxHeap(topK)

	fileDownloadCount := NewFileDownloadCount("jcenter-cache/org/apache/maven/plugins/maven-compiler-plugin/3.1/maven-compiler-plugin-3.1.pom", 3)
	maxHeap.Insert(fileDownloadCount)
	fileDownloadCount = NewFileDownloadCount("jcenter-cache/org/apache/maven/maven-artifact/2.0.8/maven-artifact-2.0.8.pom ", 3)
	maxHeap.Insert(fileDownloadCount)
	fileDownloadCount = NewFileDownloadCount("naga", 2)
	maxHeap.Insert(fileDownloadCount)
	fileDownloadCount = NewFileDownloadCount("pothula", 7)
	maxHeap.Insert(fileDownloadCount)

	fmt.Println("The Max Heap is ")
	for i := 0; i < maxHeap.size; i++ {
		fmt.Println(maxHeap.Remove())
	}
}
