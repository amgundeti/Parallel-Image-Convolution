package scheduler

import(
	"sync"
	"proj2/png"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"image"
	// "time"
)

type bspWorkerContext struct {
	//Synchronization Tools
	Cond *sync.Cond
	Mu sync.Mutex
	Done bool

	//Data for Sync
	Lead int
	NumThreads int
	FinishedThreads int
	WaitingThreads int
	DirectoryIdx int
	EffectCount int
	SectionHeight int
	Directories []string

	//Tools & data for input
	Decoder *json.Decoder
	Request *Request
	Img *png.ImageTask
}

func NewBSPContext(config Config) *bspWorkerContext {

	ctx := &bspWorkerContext{}

	ctx.Cond = sync.NewCond(&ctx.Mu)
	ctx.Done = false

	ctx.Lead = config.ThreadCount - 1
	ctx.NumThreads = config.ThreadCount
	ctx.WaitingThreads = 0

	effectsPathFile := fmt.Sprintf("../data/effects.txt")
	effectsFile, _ := os.Open(effectsPathFile)
	ctx.Decoder = json.NewDecoder(effectsFile)
	ctx.Request = &Request{Effects: []string{}}

	ctx.Directories = strings.Split(config.DataDirs, "+")
	ctx.DirectoryIdx = 0
	ctx.EffectCount = 0
	ctx.FinishedThreads = 0
	return ctx
}

func RunBSPWorker(id int, ctx *bspWorkerContext) {

	for{

		ctx.Mu.Lock()
		ctx.WaitingThreads += 1

		if ctx.WaitingThreads == ctx.NumThreads{
			ctx.WaitingThreads = 0

			if ctx.EffectCount == 0{
				// If worked through all directories for given image name, get next request
				//If not, upload image from next directory with same name
				if ctx.DirectoryIdx == 0{
					if ctx.Decoder.More(){
						ctx.Request = &Request{}
						ctx.Decoder.Decode(ctx.Request)
						filePath := "../data/in/" + ctx.Directories[ctx.DirectoryIdx] + "/"+ ctx.Request.InPath
						ctx.Img, _ = png.Load(filePath)
						ctx.SectionHeight = ctx.Img.Bounds.Max.Y / ctx.NumThreads
						ctx.Cond.Broadcast()
						ctx.Mu.Unlock()
					} else {
						ctx.Done = true
						ctx.Cond.Broadcast()
						ctx.Mu.Unlock()
						return
					}
				} else if ctx.DirectoryIdx < len(ctx.Directories){
					filePath := "../data/in/" + ctx.Directories[ctx.DirectoryIdx] + "/"+ ctx.Request.InPath
					ctx.Img, _ = png.Load(filePath)
					ctx.SectionHeight = ctx.Img.Bounds.Max.Y / ctx.NumThreads
					ctx.Cond.Broadcast()
					ctx.Mu.Unlock()
				}
			} else{
				ctx.Cond.Broadcast()
				ctx.Mu.Unlock()
			}
		} else{
			ctx.Cond.Wait()

			if ctx.Done{
				ctx.Mu.Unlock()
				return
			}
			ctx.Mu.Unlock()
		}

		YStart := id * ctx.SectionHeight
		YEnd := YStart + ctx.SectionHeight
		
		// Lead thread gets image "stub"
		if id == ctx.Lead{
			YEnd = ctx.Img.Bounds.Max.Y	
		}

		ctx.Img.Convolute(ctx.Request.Effects[ctx.EffectCount], YStart, YEnd)

		ctx.Mu.Lock()
		ctx.FinishedThreads += 1

		// If thread is last thread to show up, update indexes and save image if necessary
		if ctx.FinishedThreads == ctx.NumThreads {
			ctx.FinishedThreads = 0
			ctx.EffectCount +=1

			if ctx.EffectCount == len(ctx.Request.Effects){
				filePath := ctx.Directories[ctx.DirectoryIdx] + "_" + ctx.Request.OutPath
				ctx.Img.Save(filePath)
				ctx.EffectCount = 0
				ctx.DirectoryIdx += 1

				if ctx.DirectoryIdx == len(ctx.Directories){
					ctx.DirectoryIdx = 0
				}
			} else{
				ctx.Img.In = ctx.Img.Out
				ctx.Img.Out = image.NewRGBA64(ctx.Img.Bounds)
			}

			ctx.Cond.Broadcast()
			ctx.Mu.Unlock()

		} else{
			ctx.Cond.Wait()
			ctx.Mu.Unlock()
		}

	}
}