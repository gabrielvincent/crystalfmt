# Spaces between methods are formatted correctly
def method
end
def method_2
end
def foo
end

# Irregular spaces are handled
def method
end



def method_2
end



def foo
end

# Both classes and methods are spaced correctly
class Animal
end
def top_level_meth
end
class Person
end
def more_meth
end
def another_meth
end
class Employee
end
class Company
end

# Expressions are formatted correctly inside of classes and methods
class Animal
    val1 = 2
    val2 = 1

    def method_1
    end

    def method_2
    end
end

def method_1
    puts "yo"

    puts "gimme some space"
    puts "but not so much"
end

