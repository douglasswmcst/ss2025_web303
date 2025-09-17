import React, { useState, useEffect } from "react";
import "./App.css";

interface FoodItem {
  id: string;
  name: string;
  price: number;
}

interface Order {
  id: string;
  item_ids: string[];
  status: string;
}

function App() {
  const [items, setItems] = useState<FoodItem[]>([]);
  const [cart, setCart] = useState<FoodItem[]>([]);
  const [message, setMessage] = useState("");

  useEffect(() => {
    // We fetch from the API Gateway's route, not the service directly
    fetch("/api/catalog/items")
      .then((res) => res.json())
      .then((data) => setItems(data))
      .catch((err) => console.error("Error fetching items:", err));
  }, []);

  const addToCart = (item: FoodItem) => {
    setCart((prevCart) => [...prevCart, item]);
  };

  const placeOrder = () => {
    if (cart.length === 0) {
      setMessage("Your cart is empty!");
      return;
    }

    const order = {
      item_ids: cart.map((item) => item.id),
    };

    fetch("/api/orders/orders", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(order),
    })
      .then((res) => res.json())
      .then((data: Order) => {
        setMessage(`Order ${data.id} placed successfully!`);
        setCart([]); // Clear cart
      })
      .catch((err) => {
        setMessage("Failed to place order.");
        console.error("Error placing order:", err);
      });
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>Student Cafe</h1>
      </header>
      <main className="container">
        <div className="menu">
          <h2>Menu</h2>
          <ul>
            {items.map((item) => (
              <li key={item.id}>
                <span>
                  {item.name} - ${item.price.toFixed(2)}
                </span>
                <button onClick={() => addToCart(item)}>Add to Cart</button>
              </li>
            ))}
          </ul>
        </div>
        <div className="cart">
          <h2>Your Cart</h2>
          <ul>
            {cart.map((item, index) => (
              <li key={index}>{item.name}</li>
            ))}
          </ul>
          <button onClick={placeOrder} className="order-btn">
            Place Order
          </button>
          {message && <p className="message">{message}</p>}
        </div>
      </main>
    </div>
  );
}

export default App;
