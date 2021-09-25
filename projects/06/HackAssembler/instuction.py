class AInstruction:
    def __init__(self, symbol):
        self.symbol = symbol


class CInstruction:
    def __init__(self, dest, comp, jump):
        self.dest = dest
        self.comp = comp
        self.jump = jump


class Label:
    def __init__(self, symbol):
        self.symbol = symbol
