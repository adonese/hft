package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

/*
 Run the matching engine for a list of input operations and returns the trades and orderbooks in a
 csv-like format. Every command starts with either "INSERT", "UPDATE" or "CANCEL" with additional
 data in the columns after the command.

 In case of insert the line will have the format:
 INSERT,<order_id>,<symbol>,<side>,<price>,<volume>
 e.g. INSERT,4,FFLY,BUY,23.45,12

 In case of update the line will have the format:
 UPDATE,<order_id>,<price>,<volume>
 e.g. UPDATE,4,23.12,11

 In case of cancel the line will have the format:
 CANCEL,<order_id>
 e.g. CANCEL,4

 Side will always be "BUY" or "SELL".
 A price is a string with a maximum of 4 digits behind the ".", so "2.1427" and "33.42" would be
 valid prices but "2.14275" would not be a valid price since it has more than 4 digits behind the
 comma.
 A volume will be an integer

 The expected output is:
 - List of trades in chronological order with the format:
   <symbol>,<price>,<volume>,<taker_order_id>,<maker_order_id>
   e.g. FFLY,23.55,11,4,7
   The maker order is the one being removed from the order book, the taker order is the incoming one matching it.
 - Then, per symbol (in alphabetical order):
   - separator "===<symbol>==="
   - bid and ask price levels (sorted best to worst by price) for that symbol in the format:
     SELL,<ask_price>,<ask_volume>
     SELL,<ask_price>,<ask_volume>
     BUY,<bid_price>,<bid_volume>
     BUY,<bid_price>,<bid_volume>
     e.g. SELL,25.67,102
          SELL,25.56,34
          BUY,25.52,23
          BUY,25.51,11
          BUY,25.43,4
*/

func main() {

	reader := bufio.NewReaderSize(os.Stdin, 16*1024*1024)

	stdout, err := os.Create(os.Getenv("OUTPUT_PATH"))
	checkError(err)

	defer stdout.Close()

	writer := bufio.NewWriterSize(stdout, 16*1024*1024)

	operationsCount, err := strconv.ParseInt(strings.TrimSpace(readLine(reader)), 10, 64)
	checkError(err)

	var operations []string

	for i := 0; i < int(operationsCount); i++ {
		operationsItem := readLine(reader)
		operations = append(operations, operationsItem)
	}

	result := runMatchingEngine(operations)

	for i, resultItem := range result {
		fmt.Fprintf(writer, "%s", resultItem)

		if i != len(result)-1 {
			fmt.Fprintf(writer, "\n")
		}
	}

	fmt.Fprintf(writer, "\n")

	writer.Flush()
}

func readLine(reader *bufio.Reader) string {
	str, _, err := reader.ReadLine()
	if err == io.EOF {
		return ""
	}

	return strings.TrimRight(string(str), "\r\n")
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
