package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type btree struct {
	val    string
	weight int
	left   *btree
	rigth  *btree
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	letterCount := make(map[string]int)

	dirPath, err := os.Getwd()
	check(err)

	fileArg := os.Args[1]
	filePath := filepath.Join(dirPath, fileArg)

	fIn, err := os.Open(filePath)
	check(err)

	defer fIn.Close()

	buffer := make([]byte, 1)

	outputName := "output.bin"
	writePath := filepath.Join(dirPath, outputName)
	fOut, err := os.Create(writePath)
	check(err)

	defer fOut.Close()

	writer := bufio.NewWriter(fOut)

	for {
		_, err = fIn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				check(err)
			}

		}

		if !isPrintableCharacter(buffer[0]) {
			continue
		}

		letter := string(buffer[0])
		letterCount[letter]++

	}

	var treeRoot btree
	codeMap := mapLetterTree(letterCount, &treeRoot)
	serializedTree := serializeTree(&treeRoot)

	fIn.Seek(0, 0)

	var binString strings.Builder
	//adiciona uma representacao serializada da arvore binaria como header do arquivo
	binString.WriteString(buildHeader(serializedTree))

	for {
		_, err = fIn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				check(err)
			}

		}

		if !isPrintableCharacter(buffer[0]) {
			continue
		}

		letter := string(buffer[0])

		code := codeMap[letter]
		binString.WriteString(code)

	}

	compressedArr := getByteArr(binString.String())
	for _, byte := range compressedArr {
		err = binary.Write(writer, binary.BigEndian, byte)
		check(err)
	}

	err = writer.Flush()
	check(err)
}

func insertIntoStack(nodeStack []btree, node btree) []btree {
	if len(nodeStack) == 0 {
		nodeStack = append(nodeStack, node)
		return nodeStack
	}

	low := 0
	high := len(nodeStack) - 1
	mid := (high + low) / 2

	for {
		if low >= high {
			if node.weight >= nodeStack[mid].weight {
				nodeStack = append(nodeStack, btree{})
				copy(nodeStack[mid+1:], nodeStack[mid:])
				nodeStack[mid] = node
				return nodeStack
			} else {
				nodeStack = append(nodeStack, btree{})
				copy(nodeStack[mid+2:], nodeStack[mid+1:])
				nodeStack[mid+1] = node
				return nodeStack
			}

		} else if nodeStack[mid].weight > node.weight {
			low = mid + 1
			mid = (high + low) / 2
		} else {
			high = mid - 1
			mid = (high + low) / 2
		}

	}
}

func mapLetterTree(letterArr map[string]int, treeRoot *btree) map[string]string {

	nodeStack := make([]btree, 0)

	//cria uma priority queue em ordem crescente de ocorrencias de letras
	for symbol, weight := range letterArr {
		node := btree{
			val:    symbol,
			weight: weight,
		}

		nodeStack = insertIntoStack(nodeStack, node)
	}

	//da "pop" nos nodes e cria novos nodes compostos dos nodes originais, adicionando eles na binary tree e insere eles novamente na queue
	//ate que nao reste nenhum node na queue
	for {
		if len(nodeStack) <= 1 {
			*treeRoot = nodeStack[0]
			break
		}
		stackTop := len(nodeStack) - 1
		leftNode, rightNode := nodeStack[stackTop], nodeStack[stackTop-1]
		nodeStack = nodeStack[:stackTop-1]

		parentNode := btree{
			val:    leftNode.val + rightNode.val,
			weight: leftNode.weight + rightNode.weight,
			left:   &leftNode,
			rigth:  &rightNode,
		}

		nodeStack = insertIntoStack(nodeStack, parentNode)
	}

	symbolCodeMap := make(map[string]string)
	currCode := ""
	getSymbolCodeArr(currCode, treeRoot, symbolCodeMap)

	return symbolCodeMap
}

func getSymbolCodeArr(currCode string, node *btree, codeMap map[string]string) {

	//base case
	if node.left == nil && node.rigth == nil {
		codeMap[node.val] = currCode
		return
	}

	//go left
	if node.left != nil {
		getSymbolCodeArr(currCode+"0", node.left, codeMap)
	}

	//go right
	if node.rigth != nil {
		getSymbolCodeArr(currCode+"1", node.rigth, codeMap)
	}
}

func getByteArr(bitString string) []byte {

	var byteArr []byte
	for i := 0; i < len(bitString); i += 8 {
		end := i + 8

		end = min(end, len(bitString))

		var b byte
		for j := i; j < i+8; j++ {
			b <<= 1
			if !(j >= end) && bitString[j] == '1' {
				b |= 1
			}
		}
		byteArr = append(byteArr, b)
	}

	return byteArr

}

func isPrintableCharacter(asciiVal uint8) bool {

	if asciiVal < 32 {
		return false
	}

	return true
}

func serializeTree(treeNode *btree) string {
	if treeNode == nil {
		return ""
	}

	stringRepresentation := "\\," + treeNode.val

	if treeNode.left == nil {
		stringRepresentation += "\\,\\N"
	}

	if treeNode.rigth == nil {
		stringRepresentation += "\\,\\N"
	}

	return stringRepresentation + serializeTree(treeNode.left) + serializeTree(treeNode.rigth)
}

func buildHeader(headerContent string) string {
	var header strings.Builder
	headerEnd := "ENDHEADER"

	for _, r := range headerContent {
		fmt.Fprintf(&header, "%08b", int64(r))
	}

	for _, r := range headerEnd {
		fmt.Fprintf(&header, "%08b", int64(r))
	}

	return header.String()
}
