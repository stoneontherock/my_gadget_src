#!/bin/bash

function fn_print_ascii_map()
{
    local code
    for code in {65..90}
    do
        [ $(( (code-64) % 3 )) -eq 0 ] && bg="\e[7m" || bg=""
        printf "${bg}\x$(echo "obase=16;${code}" |bc) $code\e[0m\n"
    done
}

function initialize_stat()
{
    local code
    for code in {48..57} {65..90}
    do
        eval _${code}_=0
    done
}


function finalize_stat()
{
    local i count

    printf "\n\n\e[7m  CHAR  TIMES  PERCENT  \n"
    for i in {48..57} {65..90}
    do
        eval count=\${_${i}_}
        [ ${count} -eq 0 ] && continue
        printf "  \"\x$(echo "obase=16; ${i}" |bc)\"   %-5s  %-9s\n"    ${count}   $(awk 'BEGIN{printf("%.1f%%", '"${count}"'*100/'"${loop}"')}')
    done
    printf "  --------------------  \n  %-21s \e[0m\n"    "SUMARY: ${loop} times"
}

function main()
{
    local rand num code loop=0
    trap '[ ${loop} -gt 0 ] && finalize_stat; exit 0;'  SIGINT
    initialize_stat
    local str=({65..90})

    while true
    do
        rand=$(head -c 2 /dev/urandom |od -l |awk  'NR==1{print $2%26}')
        code="${str[$rand]}"
        printf "\x$(echo "obase=16;${code}" |bc):"

        read num
        let loop++ ; eval let _${code}_++

        num="${num// /}"
        num="${num:-255}"
        [[ ! "${num}" =~ ^[0-9]+$ ]] && { fn_print_ascii_map ; continue; }

        [ ${code} -ne $num ] &&  printf "\e[7;1;31m${str[rand]}\e[0m\n"
    done
}

main
exit $?
