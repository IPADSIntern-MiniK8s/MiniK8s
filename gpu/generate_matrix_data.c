#include <stdio.h>
#include "files.h"

#define N 10000




int main(){
  unsigned long long size = (unsigned long long)N*N*sizeof(double);
  double *a = (double*)malloc(size);
  double *b = (double*)malloc(size);
  for( int row = 0; row < N; ++row ){
    for( int col = 0; col < N; ++col ){
      a[row*N + col] = row;
      b[row*N + col] = col+2;
    }
  }
  write_values_to_file("matrix_a_data",a,size);
  write_values_to_file("matrix_b_data",b,size);

  read_values_from_file("matrix_a_data",a,size);
  read_values_from_file("matrix_b_data",b,size);

  for( int row = 0; row < N; ++row ){
    for( int col = 0; col < N; ++col ){
      if(a[row*N + col] != row ||b[row*N + col] != col+2){
        printf("error\n");
        return -1;
      }
    }
  }
  printf("generate data success\n");
}