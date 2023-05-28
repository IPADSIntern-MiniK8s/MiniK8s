# GPU

## 交我算平台

[Slurm 作业调度系统 — 上海交大超算平台用户手册 文档 (sjtu.edu.cn)](https://docs.hpc.sjtu.edu.cn/job/slurm.html)

[AI平台使用文档 — 上海交大超算平台用户手册 文档 (sjtu.edu.cn)](https://docs.hpc.sjtu.edu.cn/job/dgx.html)

[作业示例（开发者） — 上海交大超算平台用户手册 文档 (sjtu.edu.cn)](https://docs.hpc.sjtu.edu.cn/job/jobsample2.html#cuda)

`srun -p dgx2 -N 1 -n 1 --gres=gpu:1 --cpus-per-task=6 --pty /bin/bash`直接进计算节点

`sbatch xx.slurm`提交

`squeue`查看未执行

`sacct`查看任务

## NFS

### master

`apt-get install -y nfs-kernel-server`

`/etc/exports`最后`/minik8s-sharedata *(rw,sync,no_subtree_check)`

`/etc/init.d/nfs-kernel-server restart`

### node

`apt-get install -y nfs-common`

`mount master:/minik8s-sharedata /minik8s-sharedata`

master节点创建的文件/文件夹在node节点readonly

## 流程

### 主机与容器

通过查询相关资料，发现job的数据是只能通过与主机挂载目录完成的。master节点提交任务，最终跑在node节点，如何共享文件。

1. 把gpu的job yaml放在master节点，cu和Makefile放在node节点，且为该node节点加上gpu标签，yaml中的nodeselector加此标签。 这种方式把node暴露给用户，不合适。
2. 集群部署docker registry，master节点build一个镜像，然后node节点pull。这样可以获得上传的文件，但是无法把结果拿给master。
3. NFS。master和node通过nfs，共享`/minik8s-sharedata`文件夹

yaml文件挂载容器目录与主机共享文件夹，用环境变量指定参数和gpu文件

```yaml
kind: Pod
metadata:
  name: gpu-job
  namespace: gpu
spec:
  containers:
    - name: gpu-server
      image: gpu-server
      command:
        - "./job.py"
      env:
        - name: source-path
          value: /gpu
        - name: job-name
          value: gpu-matrix
        - name: partition
          value: dgx2
        - name: "N"
          value: "1"
        - name: ntasks-per-node
          value: "1"
        - name: cpus-per-task
          value: "6"
        - name: gres
          value: gpu:1
      volumeMounts:
        - name: share-data
          mountPath: /gpu
  volumes:
    - name: share-data
      hostPath:
        path: /minik8s-sharedata/gpu/matrix
```

注意这里N是个奇怪的关键字，必须加引号否则会解析错误

容器把用户提供的source文件夹（包括代码文件和Makefile）整个上传到hpc服务器，以job-name命名，并本地生成slurm也上传到这个目录。要求用户代码将需要生成的文件放入source/result中，并且一定要提前创建好，slurm会将程序的标准输入输出也放入source/result中，最后可在主机的共享文件夹中拿到结果。

用户可以灵活地自由添加任意源文件，只需要提供Makefile。规定的只有两点：

1. 用户代码需要把生成的文件存到同目录下result子目录中
2. 用户需要配置make build和make run两个命令

### 主机与hpc服务器

交大云主机`ssh-keygen -t rsa`为本机生成rsa公私钥

`ssh-copy-id stu1648@pilogin.hpc.sjtu.edu.cn`将交大云主机公钥添加到hpc登陆节点的authorized_keys中

输入密码，添加成功

之后ssh scp均无须密码，不需要把密码硬编码在代码中，安全

### 容器与hpc服务器

#### build

所有node节点提前build镜像

可以用`nerdctl tag gpu-server master:5000/gpu-server`

`nerdctl push/pull master:5000/gpu-server --insecure-registry`

containerd的镜像与docker不共享 不可以用`docker build`（或者通过docker registry中转）

如果用`nerdctl build` 需要存在`buildctl`(任意路径可调用)与`buildkitd`(提前跑的后台进程)[Releases · moby/buildkit (github.com)](https://github.com/moby/buildkit/)

需要为容器创建.ssh目录放入主机的rsa私钥和known_hosts文件，便于不需要输yes和输密码

```dockerfile
FROM ubuntu

RUN apt-get update
RUN apt-get -y install openssh-server python3-pip vim

RUN pip3 install paramiko scp

WORKDIR /root

RUN mkdir  .ssh

COPY ./id_rsa .ssh/id_rsa

COPY ./known_hosts .ssh/known_hosts

COPY ./job.py job.py

```

`nerdctl build -t gpu-server .`

`scp stu1648@data.hpc.sjtu.edu.cn:~/test_gpu.cu .`

仅使用shell命令只能做到文件传输，无法向远程主机发送需要执行的指令

解决方法是使用expect或具有ssh功能的语言编写程序

#### expect

```shell
#!/usr/bin/expect -f
spawn bash -c "ssh stu1648@pilogin.hpc.sjtu.edu.cn"

set timeout 6
expect "stu1648@pilogin*"
send "mkdir abc\r"
expect "stu1648@pilogin*"
send "exit\r"

expect eof
```

expect正如名字所表达的，必须对服务端返回值在有限范围内进行预测，这是做不到从服务端拿到sbatch的返回值job_id的，故根据sacct或squeue查看job情况也无法完成，只能实现比较简单的提交任务操作。

#### python

使用python，可以不需要提前编译

```python
#!/usr/bin/python3
import paramiko
from time import sleep
from scp import SCPClient
from os import getenv


NREAD = 100000
ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect("pilogin.hpc.sjtu.edu.cn",22,"stu1648")


job_submit_tag = "Submitted batch job"
line_finish_tag = "[stu1648@"
PENDING = "PENDING"
COMPLETED = "COMPLETED"
FAILED = "FAILED"

source_path = getenv("source-path")
job_name = getenv("job-name")
partition= getenv("partition")
N = getenv("N")
ntasks_per_node = getenv("ntasks-per-node")
cpus_per_task = getenv("cpus-per-task")
gres = getenv("gres")

if not job_name or not source_path:
    print("env error")
    exit(0)
if source_path[-1] == "/":
    # scp send whole source
    source_path = source_path[:-1]

if not partition:
    partition = "dgx2"
if not N:
    N = 1
if not ntasks_per_node:
    ntasks_per_node = 1
if not cpus_per_task:
    cpus_per_task = 6
if not gres:
    gres = "gpu:1"

def generate_slurm():
    print("generating slurm")
    with open(f"./{job_name}.slurm","w") as f:
        f.write("#!/bin/bash\n")

        f.write(f"#SBATCH --job-name={job_name}\n")
        f.write(f"#SBATCH --partition={partition}\n")
        f.write(f"#SBATCH -N {N}\n")
        f.write(f"#SBATCH --ntasks-per-node={ntasks_per_node}\n")
        f.write(f"#SBATCH --cpus-per-task={cpus_per_task}\n")
        f.write(f"#SBATCH --gres={gres}\n")


        # result must exist . is the same dir as .slurm
        f.write(f"#SBATCH --output=result/output.txt\n")
        f.write(f"#SBATCH --error=result/error.txt\n")

        f.write(f"ulimit -s unlimited\n")
        f.write(f"ulimit -l unlimited\n")

        f.write("module load gcc/8.3.0 cuda/10.1.243-gcc-8.3.0\n")

        f.write("make build\n")
        f.write("make run\n")

def upload_source():
    print("uploading source")
    scp = SCPClient(ssh.get_transport(),socket_timeout=16)
    scp.put(source_path,recursive=True,remote_path=f"~/{job_name}")
    scp.put(f"./{job_name}.slurm",f"~/{job_name}/{job_name}.slurm")
    scp.close()


def download_result(job_id):
    print("downloading result")
    scp = SCPClient(ssh.get_transport(),socket_timeout=16)
    scp = SCPClient(ssh.get_transport(),socket_timeout=16)
    #scp.get(f"~/result/{job_id}.out",f"{source_path}/{job_name}.out")
    #scp.get(f"~/result/{job_id}.err",f"{source_path}/{job_name}.err")
    scp.get(f"~/{job_name}/result",recursive=True,local_path=f"{source_path}/")
    scp.close()


def submit_job():
    t = 3
    while t:
        s = ssh.invoke_shell()
        print("starting ssh")
        sleep(2)
        recv = s.recv(NREAD).decode('utf-8')
        if recv.find("stu1648") == -1:
            print("start ssh failed,retrying")
            t -= 1
            sleep(5)
            continue

        print("start ssh success")

        print("sending sbatch")
        s.send(f"cd ~/{job_name} && sbatch ./{job_name}.slurm\n")
        sleep(5)

        recv = s.recv(NREAD).decode('utf-8')
        index = recv.find(job_submit_tag)
        if index ==-1:
            print(recv)
            print("sbatch failed,retrying")
            t -= 1
            sleep(5)
            continue
        print("sbatch success")
        job_id = recv[index+len(job_submit_tag)+1:recv.index(line_finish_tag)-2]
        #job_id = 25099457
        print(f"{job_id=}")

        print("start checking job status")

        check_status_cmd = f"sacct | grep {job_id} | awk '{{print $6}}'"

        while True:
            s.send(check_status_cmd+"\n")
            sleep(2)

            recv = s.recv(NREAD).decode('utf-8')
            status = recv[recv.index(check_status_cmd)+len(check_status_cmd)+2:recv.index(line_finish_tag)-2]
            print(f"{status=}")

            if status.find(FAILED)!=-1:
                print("job failed")
                #user might need error message, still get results
                #exit(0)
                return job_id
            if status.find(COMPLETED)==-1:
                sleep(10)
            else:
                return job_id


generate_slurm()
upload_source()
job_id =  submit_job()
if job_id:
    download_result(job_id)
print("finish")

```

由于containerd目前的使用是直接把所有容器的输出定为stdout，并没有用类似nerdctl支持-d和logs。这样会把python程序的输出都直接放到屏幕上，这样也方便观察结果，可以认为是个feature。

如果源数据文件比较大，scp上传需要花不少时间

## CUDA程序

交我算只适合提交自己确认正确的程序，因为几乎全是pending，完全无法动态调试。每天晚上十一点之后会稍微好一点。

[Google Colab](https://colab.research.google.com/github/hussain0048/C-Plus-Plus/blob/master/Basic_of_C%2B%2B.ipynb)用这个可以有cuda环境来调试

由于希望支持的GPU应用是有意义的，所以需要满足两点：

1. 数据量大
2. 由用户自己提供输入，以文件形式获得输出

故添加文件输入输出模块，并提前准备好测试数据。

为了编译任务快速结束以便展示并且减轻交我算平台的压力，这里还是选择只使用10000*10000的矩阵

```c
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
```



### 矩阵加法

```c
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

```

### 矩阵乘法

```c
#include <stdio.h>
#include <files.h>
#define CHECK_CORRECTNESS

#define N  10000

__global__ void matrixMulGPU( double * a, double * b, double * c )
{

  int row_begin = blockIdx.x * blockDim.x + threadIdx.x;
  int col_begin = blockIdx.y * blockDim.y + threadIdx.y;
  int stride_row = gridDim.x * blockDim.x;
  int stride_col = gridDim.y * blockDim.y;

  for(int row = row_begin; row < N ;row += stride_row) {
        for(int col= col_begin; col< N ; col+= stride_col) {
                double val = 0;
                for(int k = 0; k < N; ++k ){
                        val += a[row * N + k] * b[k * N + col];
                        c[row * N + col] = val;
                }
        }
  }
}

void matrixMulCPU( double * a, double * b, double * c )
{

  for( int row = 0; row < N; ++row )
    for( int col = 0; col < N; ++col )
    {
      double val = 0;
      for ( int k = 0; k < N; ++k )
        val += a[row * N + k] * b[k * N + col];
      c[row * N + col] = val;
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
  matrixMulGPU <<< number_of_blocks, threads_per_block >>> ( a, b, c_gpu );
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
    matrixMulCPU( a, b, c_cpu );
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

```

### 编译

```makefile
build:
        nvcc -o matrix-mul matrix-mul.cu -I .
run:
        ./matrix-mul

```

-I 表示在当前目录寻找include的头文件
