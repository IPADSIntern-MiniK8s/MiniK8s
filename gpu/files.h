#include <stdio.h>
#include <math.h>
#include <iostream>
#include <fstream>
#include <cstdint>
#include <string>

void read_values_from_file(const char * file, double * data, size_t size) {
    std::ifstream values(file, std::ios::binary);
    values.read(reinterpret_cast<char*>(data), size);
    values.close();
}

void write_values_to_file(const char * file, double * data, size_t size) {
    std::ofstream values(file, std::ios::binary);
    values.write(reinterpret_cast<char*>(data), size);
    values.close();
}
