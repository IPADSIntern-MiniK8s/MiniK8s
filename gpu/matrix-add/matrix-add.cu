#include <stdio.h>
#include <files.h>
#define CHECK_CORRECTNESS

#define N  10000

__global__ void matrixAddGPU( double * a, double * b, double * c )
{

  int row_begin = blockIdx.x * blockDim.x + threadIdx.x;
  int col_begin = blockIdx.y * blockDim.y + threadIdx.y;
  int stride_row = gridDim.x * blockDim.x;
  int stride_col = gridDim.y * blockDim.y;

  for(int row = row_begin; row < N ;row += stride_row) {
        for(int col= col_begin; col< N ; col+= stride_col) {
                c[row * N + col] = a[row*N+col] + b[row*N+col];
        }
  }
}

void matrixAddCPU( double * a, double * b, double * c )
{

  for( int row = 0; row < N; ++row )
    for( int col = 0; col < N; ++col )
    {
      c[row * N + col] = a[row*N+col]+b[row*N+col];
    }
}

int main()
{
        cudaError_t cudaStatus;

  int deviceId;
  int numberOfSMs;

  cudaGetDevice(&deviceId);
  cudaDeviceGetAttribute(&numberOfSMs, cudaDevAttrMultiProcessorCount, deviceId);
  printf("SM:%d\n",numberOfSMs);//80

  double *a, *b, *c_gpu;

  unsigned long long size = (unsigned long long)N * N * sizeof (double); // Number of bytes of an N x N matrix

  // Allocate memory
  cudaMallocManaged (&a, size);
  cudaMallocManaged (&b, size);
  cudaMallocManaged (&c_gpu, size);
  read_values_from_file("matrix_a_data", a, size);
  read_values_from_file("matrix_b_data", b, size);

  //if too large,invalid configuration argument
  dim3 threads_per_block(32,32,1);
  dim3 number_of_blocks (16*numberOfSMs,16*numberOfSMs, 1);
  cudaMemPrefetchAsync(a, size, deviceId);
  cudaMemPrefetchAsync(b, size, deviceId);
  cudaMemPrefetchAsync(c_gpu, size, deviceId);
  matrixAddGPU <<< number_of_blocks, threads_per_block >>> ( a, b, c_gpu );
        cudaStatus = cudaGetLastError();
        if (cudaStatus != cudaSuccess) {
                fprintf(stderr, "call matrixAddGPU error: %s\n", cudaGetErrorString(cudaStatus));
                return -1;
        }

  cudaDeviceSynchronize(); // Wait for the GPU to finish before proceeding

  // Call the CPU version to check our work
    // Compare the two answers to make sure they are equal
  bool error = false;
  #ifdef CHECK_CORRECTNESS
    double *c_cpu;
    cudaMallocManaged (&c_cpu, size);
    matrixAddCPU( a, b, c_cpu );
    for( int row = 0; row < N && !error; ++row )
      for( int col = 0; col < N && !error; ++col )
        if (c_cpu[row * N + col] != c_gpu[row * N + col])
        {
          printf("FOUND ERROR at c[%d][%d]\n", row, col);
          error = true;
          break;
        }
    cudaFree( c_cpu );
  #endif
  if (!error)
    printf("Success!\n");
  write_values_to_file("result/matrix_c_data", c_gpu, size);
  // Free all our allocated memory
  cudaFree(a);
  cudaFree(b);
  cudaFree( c_gpu );
}
