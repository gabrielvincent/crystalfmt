# A simple comment

# Empty comment below
#

#Comment without leading space

vasco = "da gama" # inline comment

def my_awesome_meth
end # inline comment after method definition


class Animal
end # inline comment after class definition

def more_meth
    # comment inside method body
end

# Comments stick to the top of method definitions
def meth_meth
end

# But the also don't, if you don't want them to

def free_meth
end


#Same with classes
class NotAMeth
end

#See?

class This < NotAMeth
end

# Multiline comments group together
# In as many lines as yo wish
# Like a terrible poem
# As are all poems, really

# Now I'm a lonely comment

# And once again we're back to writing poetry
# This is unadvised
# To treat poetry as art

# Multiline comments group together
# In as many lines as yo wish
# Like a terrible poem
# As are all poems, really

# Now I'm a lonely comment

# And once again we're back to writing poetry
# This is unadvised
# To treat poetry as art

# We can also space sections of a comment with empty comments:
#
# This is one space. Now let me show you two:
#
#
# Cool, huh?

# Inline comments in expressions
[[1,2,3]].each do |(x, y, z)|
x #=> 1
y #=> 2
z #=> 3
end
