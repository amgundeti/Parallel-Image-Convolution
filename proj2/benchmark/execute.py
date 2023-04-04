`import subprocess
import seaborn as sns
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker
import pandas as pd

directories = ["small", "mixture", "big"]
cores = [2,4,6,8,12]

seq_data = {"small": [ ], "mixture": [ ], "big": []}

for dir in directories:
    for i in range(0,5):
        terminalcommand = f"go run ../editor/editor.go {dir}"
        process = subprocess.Popen(terminalcommand, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)
        stdout, stderr = process.communicate()
        time = float(stdout.decode().strip())
        seq_data[dir].append(time)
        print(f"test: {dir}- time: {time}")

for key in seq_data:
    avg = sum(seq_data[key])/len(seq_data[key])
    seq_data[key] = avg

# Pipeline Test
pipleine_data = {"directory": [ ], "cores": [ ], "time": [ ]}

for dir in directories:
    for core in cores:
        for i in range(0,5):
            terminalcommand = f"go run ../editor/editor.go {dir} pipeline {core}"
            process = subprocess.Popen(terminalcommand, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)
            stdout, stderr = process.communicate()
            time = float(stdout.decode().strip())
            
            pipleine_data["directory"].append(dir)
            pipleine_data["cores"].append(core)
            speed_up = seq_data[dir]/time
            print(f"test: {dir} - Cores: {core} - time: {time}")
            pipleine_data["time"].append(speed_up)


df = pd.DataFrame(pipleine_data)

avg_time = df.groupby(['directory', 'cores']).mean().reset_index()

for dir in avg_time['directory'].unique():
    df_type = avg_time[avg_time['directory'] == dir]
    plt.plot(df_type['cores'], df_type['time'], label=dir)

plt.xlabel('Number of Cores')
plt.ylabel('Speed Up')
plt.title("Editor Speed Up Graph (Pipeline)")
plt.legend()
plt.savefig("speedup-pipeline1.png")
plt.close()


# BSP
bsp_data = {"directory": [ ], "cores": [ ], "time": [ ]}

for dir in directories:
    for core in cores:
        for i in range(0,5):
            terminalcommand = f"go run ../editor/editor.go {dir} pipeline {core}"
            process = subprocess.Popen(terminalcommand, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True)
            stdout, stderr = process.communicate()
            time = float(stdout.decode().strip())
            
            bsp_data["directory"].append(dir)
            bsp_data["cores"].append(core)
            speed_up = seq_data[dir]/time
            print(f"test: {dir} - Cores: {core} - time: {time}")
            bsp_data["time"].append(speed_up)


df = pd.DataFrame(bsp_data)

avg_time1 = df.groupby(['directory', 'cores']).mean().reset_index()

for dir in avg_time1['directory'].unique():
    df_type = avg_time1[avg_time1['directory'] == dir]
    plt.plot(df_type['cores'], df_type['time'], label=dir)

plt.xlabel('Number of Cores')
plt.ylabel('Speed Up')
plt.title("Editor Speed Up Graph (BSP)")
plt.legend()
plt.savefig("speedup-bsp1.png")



`