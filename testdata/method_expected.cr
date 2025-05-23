# Parameter-less method
def method
    puts "look, ma, no params"
end

# Method with params aplenty
def params_galore(one_param, two_param, three_param)
    # nothing to be done
end

def method_with_multiple_indentation_levels
    if this_level?
        if next_level?
            # I could continue all day...
        end
    end
end

def method_with_empty_parentheses()
    # I fear nothing
end

def another_one_liner(param)
    param.upcase
end

def yet_another(a, b)
    a + b
end

def with_defaults(regular, named_param = "default", another = 123)
    puts regular
    puts named_param
    puts another
end

with_defaults("regular", named_param: "non-default value", another: 1)

def messy_defaults(a, b = 1, c = "hello", d = 42)
    # this is intentionally messy
end

def with_block(&block)
    yield if 0 == 0
    puts "Done with block"
end

with_block do
    puts "grila"
end

def messy_yield
    yield(1, 2, 3)
    yield
end

foo bar do
    something
end

# The above is the same as
foo(bar) do
    something
end

foo bar { something }

# The above is the same as
foo(bar { something })

twice() do
    puts "Hello!"
end

twice do
    puts "Hello!"
end

twice { puts "Hello!" }

open file "foo.cr" do
    something
end

# Same as:
open(file("foo.cr")) do
    something
end

# Block with params
twice do |i|
    puts "Got #{i}"
end

# Scurlyful block with params
twice { |i| puts "Got #{i}" }

# Many values yield
def many(&)
    yield 1, 2, 3
end

# Many params block
many do |x, y, z|
    puts x + y + z
end

# Params with underscore
pairs do |_, second|
    print second
end

# Short one-param
method &.some_method

method(&.some_method)

["a", "b"].join(",", &.upcase)

["a", "b"].join(",") { |s| s.upcase }

# With arguments
["i", "o"].join(",", &.upcase(Unicode::CaseOptions::Turkic))

# With operators
method &.+(2)

method(&.[index])

# Yield value
def twice(&)
    v1 = yield 1
    puts v1

    v2 = yield 2
    puts v2
end

twice do |i|
    i + 1
end

# Type restrictions
def transform_int(start : Int32, &block : Int32 -> Int32)
    result = yield start
    result * 2
end

# Break in block call
thrice do |i|
    if i == 2
        break
    end
end

# Block result assigned to variable
value = twice do |i|
    if i == 1
        break "hello"
    end
    i + 1
end

# Scurlyful, multi-line block
twice { |i|
    if i == 1
        break "hello"
    end
    i + 1
}

# With ... yield
class Foo
    def one
        1
    end

    def yield_with_self(&)
        with self yield
    end

    def yield_normally(&)
        yield
    end
end

# Unpacking block params
array = [{1, "one"}, {2, "two"}]
array.each do |(number, word)|
    puts "#{number}: #{word}"
end

# The above is simply syntax sugar of this:
array = [{1, "one"}, {2, "two"}]
array.each do |arg|
    number = arg[0]
    word = arg[1]
    puts "#{number}: #{word}"
end

# Parameter unpacking can be nested.
ary = [
    {1, {2, {3, 4}}},
]

# Tree Sitter parses parameter unpacking as error
ary.each do |(w, (x, (y, z)))|
    w # => 1
    x # => 2
    y # => 3
    z # => 4
end

# Splat parameters are supported.
ary = [
    [1, 2, 3, 4, 5],
]

ary.each do |(x, *y, z)|
    x # => 1
    y # => [2, 3, 4]
    z # => 5
end
