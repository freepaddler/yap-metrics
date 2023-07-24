client.test("400 Bad Request", function () {
    client.assert(response.status === 400, "Response status is not 400 Bad Request")
})