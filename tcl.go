// Contains all the long tcl scripts
package main

const newShell = `
grid [text %{shell}] -row 0 -column 0 -sticky nwes -columnspan 2
%{shell} configure -highlightthickness 0
%{shell} configure -wrap %{wrap}
%{shell} configure -insertofftime 0
%{shell} configure -bg %{shell-bg} -fg %{shell-fg}
%{shell} configure -font %{font%q}
%{shell} configure -padx %{shell-padding}
%{shell} configure -pady %{shell-padding}
%{shell} configure -insertbackground %{shell-cursor}
bind %{shell} <Control-KeyPress-n> {edi::New}

grid [label %{prompt} -text %{prompt-value%q}] -row 1 -column 0 -sticky we
%{prompt} configure -bg %{prompt-bg} -fg %{prompt-fg}
%{prompt} configure -padx 0
%{prompt} configure -pady 0
%{prompt} configure -font %{font%q}

grid [entry %{input}] -row 1 -column 1 -sticky nwes
%{input} configure -bg %{prompt-bg} -fg %{prompt-fg}
%{input} configure -highlightthickness 0
%{input} configure -insertbackground %{shell-cursor}
%{input} configure -font %{font%q}
%{input} configure -borderwidth 0
%{input} configure -takefocus 1
%{input} configure -textvariable %{inputvar}
bind %{input} <KeyPress-Return> {edi::Exec %{id}; break}
bind %{input} <Control-KeyPress-n> {edi::New}

grid columnconfigure .%{toplevel} 0 -weight 0
grid rowconfigure .%{toplevel} 0 -weight 1

grid columnconfigure .%{toplevel} 1 -weight 1
grid rowconfigure .%{toplevel} 1 -weight 0

wm title .%{toplevel} {EDI - %{id}}
focus %{input}
`
