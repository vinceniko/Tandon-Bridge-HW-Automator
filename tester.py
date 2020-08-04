from typing import Callable, Any, List, Union, Iterable

import pexpect

import sys
import os

import string


class Input:
    def __init__(self, val):
        self.val = val

    def execute(self, stream: pexpect.spawn):
        stream.sendline(str(self.val))

class Command:
    def __init__(self, fn: Callable[[pexpect.spawn], Any]):
        self.fn = fn

    def execute(self, stream: pexpect.spawn):
        self.fn(stream)


def execute(stream: pexpect.spawn, inputs: List[Union[Input, Command]]):
    # child.expect([":", pexpect.TIMEOUT], timeout=timeout)
    for group in inputs:
        for item in group:
            item.execute(stream)
        yield


def asert(r, val, output_mod: Callable[[str], str] = lambda x: x):
    def inner(stream):
        stream.expect(r)
        output = stream.after.decode("utf-8")
        output = output_mod(output)
        try:
            assert str(val) == output
        except:
            print("assertion failed")
    return inner

inputs: List[Iterable[Union[Input, Command]]] = [
    [
        Input(90), Input(60),
    ],
    [
        Input(135), Input(71),
    ],
    [
        Input(200), Input(72),
    ],
    [
        Input(400), Input(72),
    ]
]

go_cpp_compiler = ""
if sys.argv[1] != "-":
    go_cpp_compiler = sys.argv[1]
else:
    go_cpp_compiler = "/Users/vincent/Library/Mobile Documents/com~apple~CloudDocs/Tandon/TA/c_compiler/go_c_compiler.go"

cmd = f'go run "{go_cpp_compiler}" -inDir "{os.getcwd()}/Submissions" {" ".join(sys.argv[2:] + ["-times", str(len(inputs))])}'

child = pexpect.spawn(cmd)

timeout = 1

try:
    child.expect_exact(".cpp")
    child.interact(escape_character=chr(ord(' ')))
    while True:
        # print(child.before)
        for _ in execute(child, inputs):
            child.sendline()
        child.interact(escape_character=chr(ord(' ')))
        # print(child.after)
except Exception as e:
    print(str(child))
    print(e)
