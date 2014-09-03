// Contains all the long tcl scripts
package main

const newShell = `
# Shell Frame
grid [ttk::frame %{shell}] -column 0 -row 0 -sticky nwes
grid columnconfigure %{shell} 0 -weight 0
grid rowconfigure %{shell} 0 -weight 1
grid columnconfigure %{shell} 1 -weight 1
grid rowconfigure %{shell} 1 -weight 0

# Editor component
grid [text %{editor}] -row 0 -column 0 -sticky nwes -columnspan 2
%{editor} configure -highlightthickness 0
%{editor} configure -wrap %{wrap}
%{editor} configure -insertofftime 0
%{editor} configure -bg %{editor-bg} -fg %{editor-fg}
%{editor} configure -font %{font%q}
%{editor} configure -padx %{editor-padding}
%{editor} configure -pady %{editor-padding}
%{editor} configure -insertbackground %{editor-cursor}
%{editor} configure -selectbackground %{editor-sel-bg} 
bind %{editor} <Control-KeyPress-n> {edi::New %{col-id}}
bind %{editor} <Control-KeyPress-N> {edi::NewCol %{col-id}}
bind %{editor} <Control-space> {focus %{input}}

# Shell prompt label
grid [label %{prompt} -text %{ps1}] -row 1 -column 0 -sticky we
%{prompt} configure -bg %{prompt-bg} -fg %{prompt-fg}
%{prompt} configure -padx 0
%{prompt} configure -pady 0
%{prompt} configure -font %{font%q}

# Shell prompt input
grid [entry %{input}] -row 1 -column 1 -sticky nwes
%{input} configure -bg %{prompt-bg} -fg %{prompt-fg}
%{input} configure -highlightthickness 0
%{input} configure -insertbackground %{editor-cursor}
%{input} configure -font %{font%q}
%{input} configure -borderwidth 0
%{input} configure -takefocus 1
%{input} configure -textvariable %{inputvar}
bind %{input} <KeyPress-Return> {edi::Exec %{id}; break}
bind %{input} <Control-KeyPress-n> {edi::New %{col-id}}
bind %{input} <Control-KeyPress-N> {edi::NewCol %{col-id}}
focus %{input}

%{col} add %{shell} -weight 1
`
