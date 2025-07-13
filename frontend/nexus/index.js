let apiBase = "http://127.0.0.1:7212"

async function getNewMachine() {
    console.log("getnewmachine");
    let newMachineResponse;
    await fetch(apiBase + "/new-machine", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (!response.ok) {
            throw new Error("response err");
        }
        return response.json();
    }).then(data => {
        newMachineResponse = data;
    });

    console.log(newMachineResponse)

    const response = await fetch(apiBase + "/private/ssh-key", {
        method: "GET",
        headers: {
            "Authorization": newMachineResponse["token"]
        },
    });

    const blob = await response.blob();
    const blobUrl = URL.createObjectURL(blob);

    const a = document.createElement('a');
    a.href = blobUrl;
    a.download = newMachineResponse["machine_name"] + "_ssh_key"
    document.body.appendChild(a);
    a.style.display = 'none';
    a.click();
    a.remove();

    URL.revokeObjectURL(blobUrl);
}