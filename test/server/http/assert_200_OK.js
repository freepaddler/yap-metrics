client.test("200 OK", function () {
    client.assert(response.status === 200, "Response status is not 200 OK")
})