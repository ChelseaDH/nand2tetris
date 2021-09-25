from instuction import AInstruction, CInstruction, Label


def parse_line(line):
    if line.startswith('@'):
        symbol = line[1:]
        return AInstruction(symbol)

    if line.startswith('('):
        label = line[1:].replace(')', '')
        return Label(label)

    dest_split = line.split('=')
    if len(dest_split) == 2:
        dest = dest_split[0]
        jump_split = dest_split[1].split(';')
    else:
        dest = None
        jump_split = dest_split[0].split(';')

    comp = jump_split[0]
    if len(jump_split) == 2:
        jump = jump_split[1]
    else:
        jump = None

    return CInstruction(dest, comp, jump)
