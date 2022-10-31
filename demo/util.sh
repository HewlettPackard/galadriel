#!/bin/bash
 
cyan=$(tput setaf 6)
yellow=$(tput setaf 3)
green=$(tput setaf 2)
purple=$(tput setaf 5)
norm=$(tput sgr0)

one() {
    colorize $cyan "ONE.ORG" ${@:1}
}
 
two() {
    colorize $yellow "TWO.ORG" ${@:1}
}

server() {
    colorize $green "SERVER" ${@:1}
}

client() {
    colorize $purple "CLIENT" ${@:1}
}

# Usage: colorize <color> <prefix> <arguments>
colorize() {
    color=$1
    prefix=$2
    cmd=${@:3}
    eval "${cmd} | sed -e 's/^/${color}[${prefix}] /' -e 's/$/${norm}/' &"
}
