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
