import random
import string
import sys
import getopt
import re

units = {"B": 1, "KB": 2**10, "MB": 2**20, "GB": 2**30, "TB": 2**40}


def parse_size(size: str):
    size = size.upper()
    #print("parsing size ", size)
    if not re.match(r' ', size):
        size = re.sub(r'([KMGT]?B)', r' \1', size)
    number, unit = [string.strip() for string in size.split()]
    return int(float(number)*units[unit])


def random_data_4k() -> bytes:
    return bytes(''.join(random.choices(string.ascii_letters +
                                        string.digits, k=4096)), encoding='utf-8')


def random_data_10M() -> bytes:
    return bytes(''.join(random.choices(string.ascii_letters +
                                        string.digits, k=2**20*10)), encoding='utf-8')


def generate(name: str, size: int):
    print(f"generate {name} {size}B")
    flush_size = parse_size('50mb')
    wt = 0
    with open(file=name, mode='wb') as f:
        while wt < size:
            if size > flush_size:
                data = random_data_10M()
            else:
                data = random_data_4k()
            f.write(data)
            wt += len(data)
            if size > flush_size:
                print(f"progress: {wt*100/size:.1f}%")
            if wt % flush_size == 0:
                f.flush()


def main(argv):
    opts, _ = getopt.getopt(argv, "n:k:")

    if len(opts) < 0:
        print(f"{__file__} -n <filename> -k <num of kb>")
        return

    name, size = ''.join(random.choices(
        string.digits + string.ascii_lowercase, k=10)), '4kb'

    for opt, arg in opts:
        if opt in ['-n']:
            name = arg
        if opt in ['-k']:
            size = arg

    generate(name, parse_size(size))


if __name__ == '__main__':
    main(sys.argv[1:])
