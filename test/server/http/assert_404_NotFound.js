client.test("404 Not Found", function () {
    client.assert(response.status === 404, "Response status is not 404 Not Found")
})