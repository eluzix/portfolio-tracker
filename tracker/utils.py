from datetime import datetime

from rich.console import Console

console = Console(record=True)


class TerminalColors:
    HEADER = '\033[95m'
    OK_BLUE = '\033[94m'
    OK_CYAN = '\033[96m'
    OK_GREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    END = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

    @classmethod
    def color(cls, txt, start=None, end=None):
        if start is None:
            start = cls.OK_GREEN
        if end is None:
            end = cls.END
        return f'{start}{txt}{end}'


def today() -> str:
    return datetime.today().strftime('%Y-%m-%d')
