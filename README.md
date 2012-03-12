gatos
=====
gatos is an atos-like tool for Mach-O/DWARF address symbolification.

It is distributed under the BSD licence.


Usage
-----
    gatos --raddr=xxx --laddr=xxxx --macho=xxx --dsym=xxx


Example
-----
    Thread 0 Crashed:
    0   AppName                  0x000043cc 0x1000 + 13260
                                 ^          ^
                   runtime address          load address
    1   CoreFoundation           0x37d7342e 0x37d60000 + 78894
    2   UIKit                    0x351ec9e4 0x351ce000 + 125412
    3   UIKit                    0x351ec9a0 0x351ce000 + 125344
	 
    $ gatos --raddr=0x000043cc --laddr=0x1000 --macho=appname --dsym=appname.dsym
    -[CPrefsViewController pickImage:] (CPrefsViewController.mm:332)


Notes
-----
* Input files are found inside app/.dSYM bundles.
* Does not support fat binaries; lipo -thin the app/dsym before processing.
* Does not support DWARF line table; line numbers are to the function, not address.

