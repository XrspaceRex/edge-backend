


const submit = () => {


    const form = document.querySelector(".grid");

    console.log(form);

    const formData = new FormData(form);

    console.log(formData);


    new 

    fetch('http://localhost:8020/api/v1/invocations', {
        method: 'POST',
        body: formData
    })
        .then(response => response.text())
        .then(data => {
            console.log('Success:', data);
        })
        .catch(error => {
            console.error('Error:', error);
        });

}