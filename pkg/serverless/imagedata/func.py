def run(x, y):
    z = x + y
    x = x - y
    y = y - x
    print(z)
    return x, y, z