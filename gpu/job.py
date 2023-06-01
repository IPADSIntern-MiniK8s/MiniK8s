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
