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

def another_one_liner(param) param.upcase end

def yet_another(a, b) a + b end

def with_defaults(regular, named_param = "default", another = 123)
puts regular
puts named_param
puts another
end

with_defaults("regular",named_param:    "non-default value"  , another: 1)

def messy_defaults(a, b= 1, c ="hello"   , d =  42)
# this is intentionally messy
end

def with_block( &block  )
yield if 0 == 0
puts "Done with block"
end

with_block do
puts   "grila"
end

def messy_yield
yield(1  , 2, 3)
        yield
end
