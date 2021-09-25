import argparse
import re
from os import path

from instuction import AInstruction, CInstruction, Label
from parser import parse_line

COMMENT_FORMAT = "//.*"
SYMBOLS = {
    "R0": 0,
    "R1": 1,
    "R2": 2,
    "R3": 3,
    "R4": 4,
    "R5": 5,
    "R6": 6,
    "R7": 7,
    "R8": 8,
    "R9": 9,
    "R10": 10,
    "R11": 11,
    "R12": 12,
    "R13": 13,
    "R14": 14,
    "R15": 15,
    "SCREEN": 16384,
    "KBD": 24576,
    "SP": 0,
    "LCL": 1,
    "ARG": 2,
    "THIS": 3,
    "THAT": 4,
}
SYMBOL_MEM_ADDR_COUNTER = 16
C_COMP_BINARY = {
    "0": "0101010",
    "1": "0111111",
    "-1": "0111010",
    "D": "0001100",
    "A": "0110000",
    "M": "1110000",
    "!D": "0001101",
    "!A": "0110001",
    "!M": "1110001",
    "D+1": "0011111",
    "A+1": "0110111",
    "M+1": "1110111",
    "D-1": "0001110",
    "A-1": "0110010",
    "M-1": "1110010",
    "D+A": "0000010",
    "D+M": "1000010",
    "D-A": "0010011",
    "D-M": "1010011",
    "A-D": "0000111",
    "M-D": "1000111",
    "D&A": "0000000",
    "D&M": "1000000",
    "D|A": "0010101",
    "D|M": "1010101",
}
C_JUMP_BINARY = {
    "JGT": "001",
    "JEQ": "010",
    "JGE": "011",
    "JLT": "100",
    "JNE": "101",
    "JLE": "110",
    "JMP": "111",
}


def parse_file(input_file_name):
    instructions = []
    with open(input_file_name, "r") as file:
        for line in file:
            # Strip comments and whitespace
            line = re.sub(COMMENT_FORMAT, "", line).strip()
            if not line:
                continue

            instructions.append(parse_line(line))

    return instructions


def resolve_labels(instructions):
    symbols = SYMBOLS.copy()
    line_count = 0
    for instruction in instructions:
        if not isinstance(instruction, Label):
            line_count = line_count + 1
            continue

        if instruction.symbol in symbols:
            continue

        symbols[instruction.symbol] = line_count

    return symbols


def write_binary(instructions, symbols, file):
    symbol_mem_add_counter = SYMBOL_MEM_ADDR_COUNTER
    for instruction in instructions:
        if isinstance(instruction, Label):
            continue

        if isinstance(instruction, AInstruction):
            if instruction.symbol.isnumeric():
                number = int(instruction.symbol)
            elif instruction.symbol in symbols:
                number = symbols[instruction.symbol]
            else:
                symbols[instruction.symbol] = symbol_mem_add_counter
                number = symbol_mem_add_counter
                symbol_mem_add_counter = symbol_mem_add_counter + 1

            file.write("0" + "{0:015b}".format(number) + "\n")

        if isinstance(instruction, CInstruction):
            prefix = "111"
            comp = C_COMP_BINARY[instruction.comp]
            dest = convert_c_instruction_dest_to_binary(instruction.dest)
            jump = convert_c_instruction_jump_to_binary(instruction.jump)

            file.write(prefix + comp + dest + jump + "\n")


def convert_c_instruction_dest_to_binary(dest):
    if dest is None:
        return "000"

    a = "1" if "A" in dest else "0"
    d = "1" if "D" in dest else "0"
    m = "1" if "M" in dest else "0"

    return a + d + m


def convert_c_instruction_jump_to_binary(jump):
    return C_JUMP_BINARY[jump] if jump else "000"


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("file", help="File to be assembled.")

    args = parser.parse_args()
    filename = path.normpath(args.file)

    instructions = parse_file(filename)
    symbols = resolve_labels(instructions)

    input_filename, _ = path.splitext(filename)
    output_filename = input_filename + ".hack"
    with open(output_filename, "w") as file:
        write_binary(instructions, symbols, file)
