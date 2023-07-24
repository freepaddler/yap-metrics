client.test("405 Method Not Allowed", function () {
    client.assert(response.status === 405, "Response status is not 405 Method Not Allowed")
})