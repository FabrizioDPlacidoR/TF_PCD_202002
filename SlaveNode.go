package main

import (
  "encoding/csv"
  "fmt"
  "bufio"
  "io"
  "log"
  "os"
  "strconv"
  "math"
  "math/rand"
  "time"
  //"io/ioutil"
  "strings"
  "encoding/json"
  "net"
  "sync"
)

var (
	wg sync.WaitGroup
	fwchan chan string
)

type Planta struct {
	S_lenght float64 `json:"s_lenght"`
	S_width float64 `json:"s_width"`
	P_lenght float64 `json:"p_lenght"`
	P_width float64 `json:"p_width"`
	Plant_type string `json:"plant_type"`
}

type NeuralNetwork struct{
	mHiddenLayer []*Neural
	mInputLayer  []*Neural
	mOutputLayer []*Neural
	mWeightHidden [][]float64
	mWeightOutput [][]float64
	mLastChangeHidden [][]float64
	mLastChangeOutput [][]float64
	mOutput []float64
	mForwardDone chan bool
	mFeedbackDone chan bool
	mLearningRate float64
}

func makeMatrix(rows,colums int, value float64) [][]float64{
	mat := make([][]float64,rows)
	for i:=0;i<rows;i++{
		mat[i] = make([]float64,colums)
		for j:=0;j<colums;j++{
			mat[i][j] = value
		}
	}
	return mat
}
func argmax(A []float64) int{
	x := 0 
	v := -1.0
	for i,a := range(A){
		if a>v{
			x = i
			v = a
		}
	}
	return x
}
func randomMatrix(rows,colums int, lower, upper float64) [][]float64{
	mat := make([][]float64,rows)
	for i:=0;i<rows;i++{
		mat[i] = make([]float64,colums)
		for j:=0;j<colums;j++{
			mat[i][j] = rand.Float64()*(upper-lower) + lower
		}
	}
	return mat
}
func NewNetwork(iInputCount,iHiddenCount,iOutputCount int,learningRate float64) (*NeuralNetwork){
	iInputCount +=1
	network := &NeuralNetwork{}
	network.mOutput = make([]float64,iOutputCount)
	network.mForwardDone = make(chan bool)
	network.mFeedbackDone = make(chan bool)
	network.mInputLayer = make([]*Neural,iInputCount)
	network.mLearningRate = learningRate
	for i:=0;i<iInputCount;i++{
		network.mInputLayer[i] = NewNeural(network,0,i,1)
	}
	network.mHiddenLayer = make([]*Neural,iHiddenCount)
	for i:=0;i<iHiddenCount;i++{
		network.mHiddenLayer[i] = NewNeural(network,1,i,iInputCount)
	}
	network.mOutputLayer = make([]*Neural,iOutputCount)
	for i:=0;i<iOutputCount;i++{
		network.mOutputLayer[i] = NewNeural(network,2,i,iHiddenCount)
	}

	network.mWeightHidden = randomMatrix(iInputCount,iHiddenCount,-0.2,0.2)
	network.mWeightOutput = randomMatrix(iHiddenCount,iOutputCount,-2.0,2.0)

	network.mLastChangeHidden = makeMatrix(iInputCount,iHiddenCount,0.0)
	network.mLastChangeOutput = makeMatrix(iHiddenCount,iOutputCount,0.0)

	return network
}
func (self * NeuralNetwork) Start(){//start all the neurals in the network
	for _,n := range self.mInputLayer{
		n.start()
	}
	for _,n := range self.mHiddenLayer{
		n.start()
	}
	for _,n := range self.mOutputLayer{
		n.start()
	}
}
func (self * NeuralNetwork) Stop(){//start all the neurals in the network

	for _,n := range self.mInputLayer{
		/* close(n.mInputChan) */
		close(n.mFeedbackChan)
	}
	wg.Done()
	for _,n := range self.mHiddenLayer{
		/* close(n.mInputChan) */
		close(n.mFeedbackChan)
	}
	wg.Done()
	for _,n := range self.mOutputLayer{
		/* close(n.mInputChan) */
		close(n.mFeedbackChan)
	}
	wg.Done()
	/* close(self.mForwardDone) */
	close(self.mFeedbackDone)
	wg.Done()
}
func (self * NeuralNetwork) Forward(input []float64 ) (output []float64){
	go func(){
		for i:=0;i<len(self.mInputLayer)-1;i++{
			self.mInputLayer[i].mInputChan <- input[i]
		}
		self.mInputLayer[len(self.mInputLayer)-1].mInputChan  <- 1.0 //bias node
	}()
	for i:=0;i<len(self.mOutput);i++{
		<-self.mForwardDone
	}
	return self.mOutput[:]
}
func (self * NeuralNetwork) Feedback(target []float64) {
	go func(){
		defer func(){recover()} ()
		for i:=0;i<len(self.mOutput);i++{
			self.mOutputLayer[i].mFeedbackChan <- target[i]
		}
	}()
	for i:=0;i<len(self.mHiddenLayer);i++{
		<- self.mFeedbackDone
	}
}
func (self * NeuralNetwork) CalcError( target []float64) float64{
	errSum := 0.0
	for i:=0;i<len(self.mOutput);i++{
		err := self.mOutput[i] - target[i]
		errSum += 0.5 * err * err
	}
	return errSum
}
func genRandomIndexArray(N int) []int{
	A := make([]int,N)
	for i:=0;i<N;i++{
		A[i]=i
	}
	//randomize
	for i:=0;i<N;i++{
		j := i+int(rand.Float64() * float64 (N-i))
		A[i],A[j] = A[j],A[i]
	}
	return A
}
func (self * NeuralNetwork) Train(inputs [][]float64, targets [][]float64, iteration int) {
	test_arraykeys := genRandomIndexArray(len(inputs))
	for i:=0;i<iteration;i++{
		cur_err:=0.0
		for j:=0;j<len(inputs);j++{
			self.Forward(inputs[test_arraykeys[j]])
			self.Feedback(targets[test_arraykeys[j]])
			cur_err += self.CalcError(targets[test_arraykeys[j]])
    }
    if i%(iteration/10)==0{
      fmt.Printf("\nEpoch %vth MSE: %.5f", i, cur_err / float64(len(inputs)))
    }
  }
  fmt.Println("\nTrained.")
}
type Neural struct{
	mInputChan chan float64
	mFeedbackChan chan float64
	mInputCount int
	mLayer int
	mNo int
	mNetwork * NeuralNetwork
	mValue float64
}
func NewNeural(iNetwork *NeuralNetwork, iLayer, iNo , iInputCount int) (*Neural){
	neural := &Neural{}
	neural.mNetwork = iNetwork
	neural.mInputCount = iInputCount
	neural.mLayer = iLayer
	neural.mInputChan = make(chan float64)
	neural.mFeedbackChan = make(chan float64)
	neural.mNo = iNo
	neural.mValue = 0.0
	return neural
}
func sigmoid(X float64) float64{
  res:=(1.0 + math.Pow(math.E, -float64(X)))
  return  1.0/res
}
func dsigmoid(Y float64) float64{
  return Y * (1.0 - Y)
}
func (self *Neural) start(){
	go func(){
		defer func(){recover()} ()
		for {
			sum := 0.0
			for i:=0;i<self.mInputCount;i++{
				value := <- self.mInputChan
				sum += value
			}
			if self.mLayer==0 {
				for i:=0;i<len(self.mNetwork.mHiddenLayer);i++{
					self.mNetwork.mHiddenLayer[i].mInputChan <- sum * self.mNetwork.mWeightHidden[self.mNo][i]
				}
			}else if self.mLayer==1 {
				sum = sigmoid(sum)
				for i:=0;i<len(self.mNetwork.mOutputLayer);i++{
					self.mNetwork.mOutputLayer[i].mInputChan <- sum * self.mNetwork.mWeightOutput[self.mNo][i]
				}
			}else {
				sum = sigmoid(sum)
				self.mNetwork.mOutput[self.mNo] = sum 
				self.mNetwork.mForwardDone <- true
			}
			self.mValue = sum
		}

	}()
	go func(){
		defer func(){recover()} ()
		for{
			if self.mLayer==0{
				return
			} else if self.mLayer==1{ 
				err :=0.0
				for i:=0;i<len(self.mNetwork.mOutput);i++{
					err += <- self.mFeedbackChan
				}
				for i:=0;i<self.mInputCount;i++{
					change := err * dsigmoid(self.mValue) * self.mNetwork.mInputLayer[i].mValue
					self.mNetwork.mWeightHidden[i][self.mNo] -= (self.mNetwork.mLearningRate*change + self.mNetwork.mLearningRate*self.mNetwork.mLastChangeHidden[i][self.mNo])
					self.mNetwork.mLastChangeHidden[i][self.mNo] = change
				}
				self.mNetwork.mFeedbackDone <- true
			} else{ 
				target := <- self.mFeedbackChan
				err := self.mValue - target
				for i:=0;i<self.mInputCount;i++{
					self.mNetwork.mHiddenLayer[i].mFeedbackChan <- err * self.mNetwork.mWeightOutput[i][self.mNo]
				}
				for i:=0;i<self.mInputCount;i++{
					change := err * dsigmoid(self.mValue) * self.mNetwork.mHiddenLayer[i].mValue
					self.mNetwork.mWeightOutput[i][self.mNo] -= (self.mNetwork.mLearningRate*change + self.mNetwork.mLearningRate*self.mNetwork.mLastChangeOutput[i][self.mNo])
					self.mNetwork.mLastChangeOutput[i][self.mNo] = change
				}

			}
		}
	}()
}
func readCSV(filepath string)([][]float64,[][]float64){
  inputs := make([][]float64,0)
  targets := make([][]float64,0)
  csvfile, err := os.Open(filepath)
  if err != nil {
    log.Fatalln("Couldn't open the csv file", err)
  }
  r := csv.NewReader(csvfile)
  iristype:=0
  for i := 0; i<1000; i++ {
    record, err := r.Read()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    setal_length,_:=strconv.ParseFloat(record[0], 64)
    if setal_length==0{
        i--
        continue
    }
    setal_width,_:=strconv.ParseFloat(record[1], 64)
    petal_length,_:=strconv.ParseFloat(record[2], 64)
    petal_width,_:=strconv.ParseFloat(record[3], 64)
    if record[4]=="setosa"{
        iristype=0
    }
    if record[4]=="versicolor"{
        iristype=1
    }
    if record[4]=="virginica"{
        iristype=2
    }
    X := make([]float64,0)
    X = append(X,setal_length)
    X = append(X,setal_width)
    X = append(X,petal_length)
    X = append(X,petal_width)
    inputs = append(inputs,X)
    Y := make([]float64,3)
    Y[iristype] = 1.0
    targets = append(targets,Y)
  }
  return inputs,targets
}
var remotehost string

func enviar(plant Planta) {

	conn, _ := net.Dial("tcp", remotehost)
	defer conn.Close()

	jsonBytes, _ := json.Marshal(plant)

	fmt.Fprintf(conn, "%s\n", string(jsonBytes))
}
func manejador(con net.Conn,nn *NeuralNetwork) {
	defer con.Close()
	r := bufio.NewReader(con)

	jsonString, _ := r.ReadString('\n')
	var plant Planta
	json.Unmarshal([]byte(jsonString), &plant)
	fmt.Println("Recibido: ", plant)
	var inputs []float64
	inputs=append(inputs,plant.S_lenght)
	inputs=append(inputs,plant.S_width)
	inputs=append(inputs,plant.P_lenght)
	inputs=append(inputs,plant.P_width)
	fmt.Println(inputs)
	wg.Add(4)
	prediction := nn.Forward(inputs)
	fmt.Println("Predicción: ",prediction)
	expect:=argmax(prediction)
	output:=""
	if expect==0{
		output="setosa"
	}
	if expect==1{
		output="versicolor"
	}
	if expect==2{
		output="virginica"
	}
	plant.Plant_type=output
	enviar(plant)
}
func enviarError(rate string){
	fmt.Println("Error: ",rate)
	conn,_ :=net.Dial("tcp","localhost:8000")
	defer conn.Close()
	fmt.Println("Enviando error a localhost:8000")
	fmt.Fprint(conn,rate)
}
func main(){
	wg.Add(4)
	nn := NewNetwork(4,4,3,0.03)
	start := time.Now()
	inputs,targets:=readCSV("iris.csv")
	train_inputs := make([][]float64,0)
	train_targets := make([][]float64,0)
	test_inputs := make([][]float64,0)
	test_targets := make([][]float64,0)
	chance:=0.7
	rand.Seed(time.Now().UTC().UnixNano())
	for i:= range inputs{
		rand:=rand.Float64()
		if rand>chance{
		test_inputs = append(test_inputs, inputs[i])
		test_targets = append(test_targets,targets[i])
		}else{
		train_inputs = append(train_inputs, inputs[i])
		train_targets = append(train_targets,targets[i])
		}
	}
	nn.Start()
	nn.Train(train_inputs,train_targets,100)
	nn.Stop()
	err_count := 0.0
	fmt.Println("Inputs\t\t\tExpected\tOutput")
	test_keys := genRandomIndexArray(len(test_inputs))
	for i:=0;i<len(test_inputs);i++{
		output := nn.Forward(test_inputs[test_keys[i]])
		calc := argmax(output)
		expect := argmax(test_targets[test_keys[i]])
		fmt.Println(test_inputs[test_keys[i]],"\t",expect,"\t\t",calc)
		if calc!=expect{
			err_count += 1.0
		}
	}
	rate:=1.0 - err_count/float64(len(test_inputs))
	fmt.Println("Total errors / inputs:",err_count,"/",len(test_inputs))
	fmt.Printf("Success rate: %0.5f\n",rate)
	fmt.Println("Time elapsed: ",time.Since(start))
	wg.Wait()
	rIng1 := bufio.NewReader(os.Stdin)
	fmt.Print("Puerto escucha: ")
	port, _ := rIng1.ReadString('\n')
	port = strings.TrimSpace(port)
	hostname := fmt.Sprintf("localhost:%s", port)

	fmt.Print("Puerto remoto: ")
	port, _ = rIng1.ReadString('\n')
	port = strings.TrimSpace(port)
	remotehost = fmt.Sprintf("localhost:%s", port)
	error_rate := fmt.Sprintf("%f", (1.0-rate))
	enviarError(error_rate)
	ln, _ := net.Listen("tcp", hostname)
	defer ln.Close()

	for {
		con, _ := ln.Accept()
		go manejador(con,nn)
	}

}