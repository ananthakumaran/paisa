from glob import glob
import os
import openai 
import time

api_key = ""
client = openai.OpenAI(api_key=api_key)

def handle_error(component_file): 
    err_path = "/home/tuba/Desktop/enfa-paisa/enfa-muhasebe/src/lib/components/error/"

    if not os.path.exists(err_path):
        os.makedirs(err_path)

    err_file_path = os.path.join(err_path, "error.txt")

    print(f"ERROR: COMP_FILE:{component_file}")
    
    with open(err_file_path, "a") as err_file:
        err_file.write(f"{component_file}\n")
    

3
def translate_text(component_content):
    prompt = f"Translate only the english texts to Turkish in this svelte component. Here is the svelte component: {component_content}"
    chat_gpt_token = 4096

    try:
        response = client.chat.completions.create(model="gpt-4",
                                                messages=[{"role":"user", 
                                                        "content":prompt}],
                                                max_tokens=chat_gpt_token)
        
        return response

        
    except openai.RateLimitError:
        print("Rate limit exceeded. Retrying...")
        time.sleep(3)
        return translate_text(component_content)
    
    except Exception as e:
        print(f"ERROR: {component_file} - {e}")
        return "Error"        
        


if __name__ == "__main__":
    component_files = glob("/home/tuba/Desktop/enfa-paisa/enfa-muhasebe/src/lib/components/*.svelte")
    
    for component_file in component_files:    
        with open(component_file, "r") as file:

            print(f"COMPONENT_FILE: {component_file}")
            component_content = file.read()

            response = translate_text(component_content=component_content)

            if response == "Error":
                handle_error(component_file)
                continue

            else:
                msg_content = response.choices[0].message.content

                output_dir = f"/home/tuba/Desktop/enfa-paisa/enfa-muhasebe/src/lib/components/output/"
                if not os.path.exists(output_dir):
                    os.makedirs(output_dir)

                output_file_path = os.path.join(output_dir, f"{os.path.basename(component_file)}.txt")
                print(f"SUCCESS: {os.path.basename(component_file)}")

                with open(output_file_path, "w") as file:
                    file.write(msg_content) 
                print(f"SUCCESS: FILE IS WRITTEN: {os.path.basename(component_file)}")

            


