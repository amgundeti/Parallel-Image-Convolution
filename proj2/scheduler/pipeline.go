package scheduler

import(
	"proj2/png"
	"fmt"
	"os"
	"encoding/json"
	"strings"
	"image"
)

func RunPipeline(config Config) {

	numThreads := config.ThreadCount
	done := make(chan interface{})
	workerReturn := make(chan bool)
	defer close(done)
	defer close(workerReturn)

	imgStream := ImageLoader(done, config)

	for i := 0; i < numThreads; i++ {
		go Worker(done, imgStream, numThreads, config, workerReturn, i)
	}

	completedWorkers := 0

	for{
		select{
		case <-workerReturn:
			completedWorkers +=1
			if completedWorkers == numThreads{
				return
			}
		}
	}
}

///////////////////////////
//Task Generator
///////////////////////////
func ImageLoader(done <-chan interface{}, config Config) <-chan *png.ImageTask{
	
	effectsPathFile := fmt.Sprintf("../data/effects.txt")
	effectsFile, err := os.Open(effectsPathFile)
	
	if err != nil {
		panic(err)
	}

	reader := json.NewDecoder(effectsFile)
	directories := strings.Split(config.DataDirs, "+")

	imgStream := make(chan *png.ImageTask)

	go func(){
		defer close(imgStream)
		for reader.More(){
			// fmt.Println("in reader.More")
			req := Request{}
			err := reader.Decode(&req)

			if err != nil {
				print(err)
				return
			}

			for _, directory := range directories{

				filePath := "../data/in/" + directory + "/"+ req.InPath

				pngImg, err := png.Load(filePath)
				if err != nil {
					print(err)
					return
				}
				pngImg.Effects = req.Effects
				outname := directory + "_" + req.OutPath
				pngImg.OutName = outname


				select{
				case <- done:
					return
				case imgStream <- pngImg:
				}
			}
		}

	}()
	return imgStream
}

///////////////////////////
//Worker
///////////////////////////
func Worker(done <-chan interface{}, imageStream <-chan *png.ImageTask, numThreads int, config Config, workerReturn chan bool, id int){

	for img := range imageStream{

		push := make(chan string)
		signal := make(chan bool)
		stopGo := make(chan interface{})
		imgSliceStream := make(chan *png.ImageTask)

		// spawn MiniWorkers as goRoutines - pulling from imgSliceStream
		areaCovered := 0
		sectionHeight := img.Bounds.Max.Y/numThreads

		for i := 0; i < numThreads; i++{
			YStart := areaCovered
			YEnd := areaCovered + sectionHeight

			if i == numThreads - 1{
				YEnd = img.Bounds.Max.Y
			}
			areaCovered += sectionHeight

			go MiniRoutine(done, push, signal, imgSliceStream, stopGo, YStart, YEnd)
			imgSliceStream <- img
		}

		// synchronize applicaton of effects
		for e := 0; e < len(img.Effects); e++{
			img.Out = image.NewRGBA64(img.Bounds)

			effect := img.Effects[e]
			for n := 0; n < numThreads; n++{
				push <- effect
			}

			completions := 0

			for j :=0; j < numThreads; j++{
				select{
				case <-done:
					return
				case <-signal:
					completions += 1
					if completions == numThreads{
						break
					}
				}
			}
			img.In = img.Out

		}
		img.Save(img.OutName)
		for n := 0; n < numThreads; n++{
			stopGo <- true
		}
		close(push)
		close(signal)

	}

	workerReturn <- true
}

///////////////////////////
//MiniWorker
///////////////////////////
func MiniRoutine(done <-chan interface{}, push <-chan string, signal chan<- bool, imgSliceStream <-chan *png.ImageTask, stopGo chan interface{}, YStart int, YEnd int){

	imgSlice:= <-imgSliceStream

	for {
		select{
		case <-done:
			return
		case e:= <-push:
			imgSlice.Convolute(e, YStart, YEnd)
			signal <- true
		case <- stopGo:
			return
		}
	}

}

