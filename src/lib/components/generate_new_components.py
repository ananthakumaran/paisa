from glob import glob



if __name__ == "__main__":
    component_files: list = glob("./*.svelte")

    for component_file in component_files:
        component_content = None
        with open(component_file, "r") as file:
            print(file.read())
            component_content = file.read()

        # 1. import chatgpt api

        # 2. prepare the request prompt

        prompt: str = f"Translate only the english texts to Turkish in this svelte component. Here is the svelte component: {component_content}"
        # 3. request to chatgpt api to get response
        # chat_gpt_url
        chat_gpt_token = "Bearer " + "your_chat_gpt_token" 