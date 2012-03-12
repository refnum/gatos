/*	NAME:
		gatos.go

	DESCRIPTION:
		Mach-O/dwarf symbol lookup.

	COPYRIGHT:
		Copyright (c) 2012, refNum Software
		<http://www.refnum.com/>

		All rights reserved.

		Redistribution and use in source and binary forms, with or without
		modification, are permitted provided that the following conditions
		are met:

			o Redistributions of source code must retain the above
			copyright notice, this list of conditions and the following
			disclaimer.

			o Redistributions in binary form must reproduce the above
			copyright notice, this list of conditions and the following
			disclaimer in the documentation and/or other materials
			provided with the distribution.

			o Neither the name of refNum Software nor the names of its
			contributors may be used to endorse or promote products derived
			from this software without specific prior written permission.

		THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
		"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
		LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
		A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
		OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
		SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
		LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
		DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
		THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
		(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
		OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
	__________________________________________________________________________
*/
//============================================================================
//		Imports
//----------------------------------------------------------------------------
package main

import "os"
import "fmt"
import "flag"
import "path"
import "debug/macho"
import "debug/dwarf"





//============================================================================
//		Globals
//----------------------------------------------------------------------------
var gCurrentFile   string;
var gTargetAddress uint64;





//============================================================================
//		processCompileUnit : Process a compile unit.
//----------------------------------------------------------------------------
func processCompileUnit(theReader *dwarf.Reader, depth int, theEntry *dwarf.Entry) {



	// Process the entry
	gCurrentFile = theEntry.Val(dwarf.AttrName).(string);

	if (theEntry.Children) {
		processChildren(theReader, depth+1, false);
	}

	gCurrentFile = "";

}





//============================================================================
//		processSubprogram : Process a subprogram.
//----------------------------------------------------------------------------
func processSubprogram(theReader *dwarf.Reader, depth int, theEntry *dwarf.Entry) {

	var lowAddr		uint64;
	var highAddr	uint64;



	// Get the state we need
	lowVal  := theEntry.Val(dwarf.AttrLowpc);
	highVal := theEntry.Val(dwarf.AttrHighpc);
	
	if (lowVal != nil && highVal != nil) {
		lowAddr  = lowVal.(uint64);
		highAddr = highVal.(uint64);
	}



	// Check for a match
	if (gTargetAddress >= lowAddr && gTargetAddress < highAddr) {
		name := theEntry.Val(dwarf.AttrName)
		line := theEntry.Val(dwarf.AttrDeclLine)

		fmt.Printf("%v (%v:%v)\n", name, path.Base(gCurrentFile), line);
		}



	// Process the entry
    if (theEntry.Children) {
        processChildren(theReader, depth+1, false);
    }
}





//============================================================================
//		readNextEntry : Read the next entry.
//----------------------------------------------------------------------------
func readNextEntry(theReader *dwarf.Reader) *dwarf.Entry {



	// Read the entry
	theEntry, theErr := theReader.Next();
	if (theErr != nil) {
		fmt.Printf("ERROR: %v\n", theErr.String());
		theEntry = nil;
	}

	return(theEntry);
}





//============================================================================
//		processEntry : Process an entry.
//----------------------------------------------------------------------------
func processEntry(theReader *dwarf.Reader, depth int, theEntry *dwarf.Entry) {



	// Process the entry
	switch theEntry.Tag {
		case dwarf.TagCompileUnit:	processCompileUnit(theReader, depth, theEntry);
		case dwarf.TagSubprogram:	processSubprogram( theReader, depth, theEntry);
		default:
			if (theEntry.Children) {
				processChildren(theReader, depth+1, true);
			}
		}
}





//============================================================================
//		processChildren : Process an entry's children.
//----------------------------------------------------------------------------
func processChildren(theReader *dwarf.Reader, depth int, canSkip bool) {



	// Process the children
	if (canSkip) {
		theReader.SkipChildren();
	} else {
		for {
			theChild := readNextEntry(theReader);
			if (theChild == nil || theChild.Tag == 0) {
				break;
			}
			
			processEntry(theReader, depth, theChild);
		}
	}
}





//============================================================================
//		fatalError : Display an error.
//----------------------------------------------------------------------------
func fatalError(theMsg string) {


	// Show the error
	fmt.Printf("%v\n", theMsg);
	os.Exit(-1);
}





//============================================================================
//		printHelp : Print some help.
//----------------------------------------------------------------------------
func printHelp() {



	// Print some help
	fmt.Print("\n");
	fmt.Print("Usage:\n");
	fmt.Print("    gatos --raddr=xxx --laddr=xxxx --macho=xxx --dsym=xxx\n");
	fmt.Print("\n");
	fmt.Print("Example:\n");
	fmt.Print("    Thread 0 Crashed:\n");
	fmt.Print("    0   AppName                  0x000043cc 0x1000 + 13260\n");
	fmt.Print("                                 ^          ^\n");
	fmt.Print("                   runtime address          load address\n");
	fmt.Print("    1   CoreFoundation           0x37d7342e 0x37d60000 + 78894\n");
	fmt.Print("    2   UIKit                    0x351ec9e4 0x351ce000 + 125412\n");
	fmt.Print("    3   UIKit                    0x351ec9a0 0x351ce000 + 125344\n");
	fmt.Print("\n");
	fmt.Print("    $ gatos --raddr=0x000043cc --laddr=0x1000 --macho=appname --dsym=appname.dsym\n");
	fmt.Print("    -[CPrefsViewController pickImage:] (CPrefsViewController.mm:332)\n");
	fmt.Print("\n");
	fmt.Print("Notes:\n");
	fmt.Print("    o Input files are found inside app/.dSYM bundles.\n");
	fmt.Print("    o Does not support fat binaries; lipo -thin the app/dsym before processing.\n");
	fmt.Print("    o Does not support DWARF line table; line numbers are to function, not address.\n");
	fmt.Print("\n");
	
	os.Exit(-1);
}





//============================================================================
//		main : Entry point.
//----------------------------------------------------------------------------
func main() {

	var dwarfData			*dwarf.Data;
	var theFile				*macho.File;
	var theErr				os.Error;
	var relativeAddress		uint64;
	var runtimeAddress		uint64;
	var loadAddress			uint64;
	var segmentAddress		uint64;
	var pathMacho			string;
	var pathDsym			string;



	// Parse our arguments
	flag.Uint64Var(&runtimeAddress, "raddr",  0, "");
	flag.Uint64Var(&loadAddress,    "laddr",  0, "");
	flag.StringVar(&pathMacho,      "macho", "", "");
	flag.StringVar(&pathDsym,       "dsym",  "", "");
	flag.Parse();

	if (runtimeAddress == 0 || loadAddress == 0 || pathMacho == "" || pathDsym == "") {
		printHelp();
	}



	// Find the text segment address
	theFile, theErr = macho.Open(pathMacho);
	if (theErr != nil) {
		fatalError("Can't open Mach-O file: " + theErr.String());
	}

	segmentAddress = theFile.Segment("__TEXT").Addr;

	theFile.Close();



	// Calculate the target address
	relativeAddress = runtimeAddress - loadAddress;
	gTargetAddress  = segmentAddress   + relativeAddress;



	// Find the target
	theFile, theErr = macho.Open(pathDsym);
	if (theErr != nil) {
		fatalError("Can't open .dsym file: " + theErr.String());
	}

	dwarfData, theErr = theFile.DWARF();
	if (theErr != nil) {
		fatalError("Can't find DWARF info: " + theErr.String());
	}

	processChildren(dwarfData.Reader(), 0, false);

	theFile.Close();
}





