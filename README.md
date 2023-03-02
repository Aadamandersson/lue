# lue

## Installation

```shell
go install github.com/aadamandersson/lue/cmd/lue
```

## Usage

```shell
lue <path>
```

## Example

```text
fn fib(n: int): int {
    if n <= 1 {
        n
    } else {
        fib(n - 1) + fib(n - 2)
    }
}

fn main() {
    let result = fib(9)
    println(result)
}
```

More examples can be found in the tests/ folder.

## License

MIT licensed. See the [LICENSE](./LICENSE) file for details.
