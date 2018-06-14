package main

import (
	"encoding/binary"
	"time"
	"github.com/ReconfigureIO/sdaccel/xcl"
	"github.com/ReconfigureIO/fixed"
	"github.com/ReconfigureIO/fixed/host"
	"log"
	"github.com/lebu26/TestData"
)

func main(){

	var length uint32 = 1  
	
	log.Printf("**********************program svm_1 starting*********************\n\r")

	log.Printf("************************************************************************************* \n\r")
	log.Printf("                                  hardware staring\n\r")
	log.Printf("************************************************************************************* \n\r")

	// Allocate a 'world' for interacting with the FPGA
	world := xcl.NewWorld()
	defer world.Release()

	// Import the compiled code that will be loaded onto the FPGA (referred to here as a kernel)
	// Right now these two identifiers are hard coded as an output from the build process
	krnl := world.Import("kernel_test").GetKernel("reconfigure_io_sdaccel_builder_stub_0_1")
	defer krnl.Release()
	log.Printf("___________________________________World creation finished___________________________ \n\r")

	log.Printf("___________________________________Trasfer datas_____________________________________ \n\r")

	f_td := [16]fixed.Int26_6{0}
	for i:=0; i<16 ;i++{
		f_td[i] = host.I26Float64(TD.TestData(i))
	}

	// Allocate a space in the shared memory to store the data you're sending to the FPGA
	buff_in := world.Malloc(xcl.ReadOnly, uint(binary.Size(f_td)))
	defer buff_in.Free()

	// Construct an array to hold the output data from the FPGA
	output := make([]uint32, length)

	// Allocate a space in the shared memory to store the output data from the FPGA.
	outputBuff := world.Malloc(xcl.ReadWrite, uint(binary.Size(output)))
	defer outputBuff.Free()

	// Write our input data to shared memory at the address we previously allocated
	binary.Write(buff_in.Writer(), binary.LittleEndian, f_td)

	// Zero out the space in shared memory for the result from the FPGA
	binary.Write(outputBuff.Writer(), binary.LittleEndian, output)
	log.Printf("___________________________________Buffers initialization finished___________________________ \n\r")

	// Pass the pointer to the input data in shared memory as the first and second argument
	krnl.SetMemoryArg(0, buff_in)
	//Pass the pointer to the output data in shared memory as the third argument
	krnl.SetMemoryArg(1, outputBuff)

	var td_hw time.Duration


	// Pass the length of the vector as the fourth argument
	krnl.SetArg(2, length)
	log.Printf("___________________________________kernel arguments initialization finished___________________________ \n\r")

	// Run the FPGA with the supplied arguments. This is the same for all projects.
	// The arguments ``(1, 1, 1)`` relate to x, y, z co-ordinates and correspond to our current
	// underlying technology.
	log.Printf("************************************************************************************* \n\r")
	log.Printf("                                  kernel staring\n\r")
	log.Printf("************************************************************************************* \n\r")
	var t1_hw = time.Now()
	krnl.Run(1, 1, 1)

	// Read the result from shared memory. If it is zero return an error
	err := binary.Read(outputBuff.Reader(), binary.LittleEndian, output)
	if err != nil {
		log.Fatal("binary.Read failed:", err)
	}

	err_count := 0
	for w:=0 ; w<int(length); w++{
		log.Printf("** output[%d] = %v, lable[%d] = %v \n\r", w,output[w],w,TD.TestDataLable(w))
		if float64(output[w]) != TD.TestDataLable(w){
			err_count ++
		}
	} 
	var t2_hw = time.Now()
	td_hw = t2_hw.Sub(t1_hw)
	log.Printf("**********************hardware finished*********************\n\r\r\r")
	log.Printf("** takes = %v s\n\r", td_hw.Seconds())
	log.Printf("************************************************************************************* \n\r")
	log.Printf("                                  kernel finished successfully\n\r")
	log.Printf("                                  error count = %v \n\r",err_count)
	log.Printf("************************************************************************************* \n\r")

}